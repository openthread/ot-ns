#!/usr/bin/env python3
# Copyright (c) 2026, The OTNS Authors.
# All rights reserved.
#
# Redistribution and use in source and binary forms, with or without
# modification, are permitted provided that the following conditions are met:
# 1. Redistributions of source code must retain the above copyright
#    notice, this list of conditions and the following disclaimer.
# 2. Redistributions in binary form must reproduce the above copyright
#    notice, this list of conditions and the following disclaimer in the
#    documentation and/or other materials provided with the distribution.
# 3. Neither the name of the copyright holder nor the
#    names of its contributors may be used to endorse or promote products
#    derived from this software without specific prior written permission.
#
# THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS"
# AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
# IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE
# ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE
# LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR
# CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF
# SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS
# INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN
# CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE)
# ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE
# POSSIBILITY OF SUCH DAMAGE.

# Case study for multi-SSED CSL wake-up crowding.
#
# Topology:
# - 1 router/leader parent
# - N SSED children
# - identical CSL period on every child
# - low traffic: once per round, each child gets one outbound UDP trigger
#
# Outputs:
# - per-round metadata CSV
# - per-child phase observations CSV
# - per-child phase summary CSV
# - adjacent wake-gap CSV
# - observed sample interval CSV
# - text report
# - markdown overview with tables and embedded plot
# - optional PNG with phase distribution and wake-gap histogram
# - PCAP capture for deeper inspection

import argparse
import csv
import logging
import math
import os
import re
import shutil
import statistics
import sys
from dataclasses import dataclass
from pathlib import Path
from typing import Dict, List, Optional, Sequence, Tuple

REPO_ROOT = Path(__file__).resolve().parents[2]
PYLIBS_ROOT = REPO_ROOT / 'pylibs'
if str(PYLIBS_ROOT) not in sys.path:
    sys.path.insert(0, str(PYLIBS_ROOT))

from otns.cli import OTNS
from otns.cli.errors import OTNSExitedError


@dataclass(frozen=True)
class ChildNode:
    node_id: int
    label: str
    extaddr: str
    rloc16: int
    x: int
    y: int


@dataclass(frozen=True)
class RoundWindow:
    round_index: int
    start_s: float
    end_s: float


@dataclass(frozen=True)
class Observation:
    round_index: int
    child_id: int
    child_label: str
    extaddr: str
    packet_number: int
    sniff_time_s: float
    delay_from_round_start_ms: float
    phase_offset_ms: float
    csl_phase_units: int
    csl_phase_ie_ms: float
    csl_period_units: int
    frame_kind: str


@dataclass(frozen=True)
class ChildSummary:
    child_id: int
    child_label: str
    extaddr: str
    samples: int
    coverage: float
    mean_phase_ms: float
    circular_std_ms: float
    min_phase_ms: float
    max_phase_ms: float


@dataclass(frozen=True)
class WakeGap:
    round_index: int
    from_child_id: int
    from_child_label: str
    to_child_id: int
    to_child_label: str
    gap_ms: float


@dataclass(frozen=True)
class SampleInterval:
    child_id: int
    child_label: str
    interval_ms: float


OBSERVATION_TOKEN_PATTERN = re.compile(r'cw(?P<round>\d+)n(?P<child>\d+)')


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description='CSL wake-up crowding case study for 1 router + N SSED topology')
    parser.add_argument('--sim-id', type=int, default=0,
                        help='OTNS simulation id used for listen ports (default: 0)')
    parser.add_argument('--num-ssed', type=int, default=8,
                        help='Number of SSED children (default: 8)')
    parser.add_argument('--csl-period-ms', type=float, default=500.0,
                        help='CSL period in milliseconds, rounded to 160us steps (default: 500)')
    parser.add_argument('--csl-timeout', type=int, default=120,
                        help='CSL synchronized timeout in seconds (default: 120)')
    parser.add_argument('--rounds', type=int, default=48,
                        help='Number of low-traffic observation rounds (default: 48)')
    parser.add_argument('--round-interval', type=float, default=8.0,
                        help='Seconds between round starts (default: 8.0)')
    parser.add_argument('--collection-window', type=float, default=None,
                        help='Seconds to wait for each round sample window (default: max(1.5, 2.5 * period))')
    parser.add_argument('--log-poll-interval', type=float, default=0.01,
                        help='Polling interval while collecting parent UDP receive logs (default: 0.01)')
    parser.add_argument('--formation-time', type=float, default=45.0,
                        help='Maximum wait time for all children to attach and show CSL in child table (default: 45)')
    parser.add_argument('--parent-boot-time', type=float, default=10.0,
                        help='Initial settling time after adding the parent router (default: 10)')
    parser.add_argument('--udp-port', type=int, default=19788,
                        help='UDP port used for observation triggers (default: 19788)')
    parser.add_argument('--router-executable', default=None,
                        help='Optional explicit path to the FTD CLI executable for the parent router')
    parser.add_argument('--child-executable', default=None,
                        help='Optional explicit path to the MTD CLI executable for SSED children')
    parser.add_argument('--results-dir', default='results',
                        help='Directory for CSV, report, image, and pcap outputs (default: results)')
    parser.add_argument('--prefix', default=None,
                        help='Optional output filename prefix')
    parser.add_argument('--radiomodel', default='MutualInterference',
                        help='OTNS radiomodel to use (default: MutualInterference)')
    parser.add_argument('--radio-range', type=int, default=180,
                        help='Parent and child radio range in OTNS units (default: 180)')
    parser.add_argument('--child-radius', type=int, default=20,
                        help='Radius of the child ring around the parent in OTNS units (default: 20)')
    parser.add_argument('--center-x', type=int, default=320,
                        help='Topology center X position (default: 320)')
    parser.add_argument('--center-y', type=int, default=240,
                        help='Topology center Y position (default: 240)')
    parser.add_argument('--skip-plot', action='store_true',
                        help='Skip PNG generation and only write CSV/report outputs')
    parser.add_argument('--log-level', default='INFO', choices=['DEBUG', 'INFO', 'WARNING', 'ERROR'],
                        help='Python logging level (default: INFO)')
    args = parser.parse_args()

    if args.num_ssed < 2:
        raise ValueError('--num-ssed must be >= 2 to evaluate crowding')
    if args.rounds <= 0:
        raise ValueError('--rounds must be > 0')
    if args.round_interval <= 0:
        raise ValueError('--round-interval must be > 0')
    if args.csl_period_ms <= 0:
        raise ValueError('--csl-period-ms must be > 0')

    args.csl_period_us = round(args.csl_period_ms * 1000.0 / 160.0) * 160
    args.csl_period_ms = args.csl_period_us / 1000.0
    args.csl_period_s = args.csl_period_us / 1_000_000.0
    args.collection_window = args.collection_window or max(1.5, args.csl_period_s * 2.5)
    if args.collection_window <= 0:
        raise ValueError('--collection-window must be > 0')
    if args.collection_window > args.round_interval:
        raise ValueError('--collection-window must be <= --round-interval')
    if args.log_poll_interval <= 0:
        raise ValueError('--log-poll-interval must be > 0')

    logging.basicConfig(level=getattr(logging, args.log_level), format='%(asctime)s - %(levelname)s - %(message)s')
    return args


def report_prefix(args: argparse.Namespace) -> str:
    if args.prefix:
        return args.prefix
    period_ms = int(round(args.csl_period_ms))
    return f'csl_crowding_n{args.num_ssed}_p{period_ms}ms_r{args.rounds}'


def build_output_paths(args: argparse.Namespace) -> Dict[str, Path]:
    results_dir = Path(args.results_dir)
    results_dir.mkdir(parents=True, exist_ok=True)
    prefix = report_prefix(args)
    return {
        'results_dir': results_dir,
        'rounds': results_dir / f'{prefix}_rounds.csv',
        'observations': results_dir / f'{prefix}_observations.csv',
        'child_summary': results_dir / f'{prefix}_child_phases.csv',
        'wake_gaps': results_dir / f'{prefix}_wake_gaps.csv',
        'sample_intervals': results_dir / f'{prefix}_sample_intervals.csv',
        'report': results_dir / f'{prefix}_report.txt',
        'overview': results_dir / f'{prefix}_overview.md',
        'plot': results_dir / f'{prefix}_plots.png',
        'pcap': results_dir / f'{prefix}.pcap',
    }


def format_extaddr_int(value: int) -> str:
    hexstr = f'{value:016x}'
    return ':'.join(hexstr[index:index + 2] for index in range(0, 16, 2))


def child_position(index: int, total: int, radius: int, center_x: int, center_y: int) -> Tuple[int, int]:
    angle = 2.0 * math.pi * index / max(total, 1)
    x = int(round(center_x + radius * math.cos(angle)))
    y = int(round(center_y + radius * math.sin(angle)))
    return x, y


def parse_cli_table(lines: Sequence[str]) -> List[Dict[str, str]]:
    normalized_lines = [line.replace('\r', '').replace('\n', '').strip() for line in lines if line.strip()]
    if len(normalized_lines) < 3:
        return []
    header_line = normalized_lines[0]
    if not header_line.startswith('|'):
        return []

    headers = [header.strip() for header in header_line.strip('|').split('|')]
    rows: List[Dict[str, str]] = []
    for line in normalized_lines[2:]:
        if not line.startswith('|'):
            continue
        values = [value.strip() for value in line.strip('|').split('|')]
        if len(values) != len(headers):
            continue
        rows.append(dict(zip(headers, values)))
    return rows


def resolve_executable(explicit_path: Optional[str], candidates: Sequence[str]) -> Optional[str]:
    if explicit_path:
        explicit = os.path.expanduser(explicit_path)
        return explicit if os.path.isfile(explicit) else shutil.which(explicit)

    for candidate in candidates:
        found = shutil.which(candidate)
        if found:
            return found
        if os.path.isfile(candidate):
            return candidate
    return None


def resolve_node_executables(args: argparse.Namespace) -> Tuple[str, str]:
    repo_root = str(REPO_ROOT)
    router_candidates = (
        'ot-cli-ftd',
        os.path.join(repo_root, 'ot-rfsim', 'ot-versions', 'ot-cli-ftd'),
        os.path.join(repo_root, 'ot-rfsim', 'build', 'latest', 'bin', 'ot-cli-ftd'),
        os.path.join(repo_root, 'ot-rfsim', 'build', 'br', 'bin', 'ot-cli-ftd'),
        os.path.join(repo_root, 'openthread', 'build', 'simulation', 'examples', 'apps', 'cli', 'ot-cli-ftd'),
    )
    child_candidates = (
        'ot-cli-mtd',
        os.path.join(repo_root, 'ot-rfsim', 'ot-versions', 'ot-cli-mtd'),
        os.path.join(repo_root, 'ot-rfsim', 'build', 'latest', 'bin', 'ot-cli-mtd'),
        os.path.join(repo_root, 'openthread', 'build', 'simulation', 'examples', 'apps', 'cli', 'ot-cli-mtd'),
        'ot-cli-ftd',
        os.path.join(repo_root, 'ot-rfsim', 'ot-versions', 'ot-cli-ftd'),
        os.path.join(repo_root, 'ot-rfsim', 'build', 'latest', 'bin', 'ot-cli-ftd'),
    )

    router_executable = resolve_executable(args.router_executable, router_candidates)
    child_executable = resolve_executable(args.child_executable, child_candidates)

    if router_executable is None:
        raise RuntimeError(
            'Could not find an FTD CLI executable. Pass --router-executable or add ot-cli-ftd to PATH.'
        )
    if child_executable is None:
        raise RuntimeError(
            'Could not find an MTD CLI executable. Pass --child-executable or add ot-cli-mtd to PATH.'
        )

    return router_executable, child_executable


def wait_for_csl_children(ns: OTNS, parent_id: int, expected_children: int, timeout_s: float) -> List[Dict[str, str]]:
    deadline = ns.time + timeout_s
    last_rows: List[Dict[str, str]] = []
    while ns.time <= deadline + 1e-6:
        last_rows = parse_cli_table(ns.node_cmd(parent_id, 'child table'))
        if len(last_rows) == expected_children and all(row.get('CSL') == '1' for row in last_rows):
            return last_rows
        remaining = deadline - ns.time
        if remaining <= 1e-6:
            break
        ns.go(min(1.0, remaining))

    raise RuntimeError(
        f'Expected {expected_children} CSL children on parent {parent_id}, got {len(last_rows)}: {last_rows}'
    )


def build_topology(ns: OTNS, args: argparse.Namespace) -> Tuple[int, List[int]]:
    ns.radiomodel = args.radiomodel
    ns.packet_loss_ratio = 0.0
    router_executable, child_executable = resolve_node_executables(args)

    parent_id = ns.add(
        'router',
        x=args.center_x,
        y=args.center_y,
        radio_range=args.radio_range,
        executable=router_executable,
    )
    ns.go(args.parent_boot_time)

    child_ids: List[int] = []
    for index in range(args.num_ssed):
        x, y = child_position(index, args.num_ssed, args.child_radius, args.center_x, args.center_y)
        child_id = ns.add(
            'ssed',
            x=x,
            y=y,
            radio_range=args.radio_range,
            executable=child_executable,
        )
        ns.node_cmd(child_id, f'csl period {args.csl_period_us}')
        ns.node_cmd(child_id, f'csl timeout {args.csl_timeout}')
        child_ids.append(child_id)

    wait_for_csl_children(ns, parent_id, args.num_ssed, args.formation_time)
    ns.node_cmd(parent_id, 'udp open')
    ns.node_cmd(parent_id, f'udp bind :: {args.udp_port}')
    for child_id in child_ids:
        ns.node_cmd(child_id, 'udp open')
    return parent_id, child_ids


def collect_child_nodes(ns: OTNS, child_ids: Sequence[int]) -> List[ChildNode]:
    node_info = ns.nodes()
    children: List[ChildNode] = []
    for node_id in child_ids:
        info = node_info[node_id]
        children.append(
            ChildNode(
                node_id=node_id,
                label=f'SSED_{node_id}',
                extaddr=format_extaddr_int(info['extaddr']),
                rloc16=info['rloc16'],
                x=info['x'],
                y=info['y'],
            )
        )
    return children


def parse_observation_token(line: str) -> Optional[Tuple[int, int]]:
    match = OBSERVATION_TOKEN_PATTERN.search(line)
    if not match:
        return None
    return int(match.group('round')), int(match.group('child'))


def run_rounds(ns: OTNS, destination: str, children: Sequence[ChildNode],
               args: argparse.Namespace) -> Tuple[List[RoundWindow], List[Observation]]:
    rounds: List[RoundWindow] = []
    observations: List[Observation] = []
    child_by_id = {child.node_id: child for child in children}

    for round_index in range(1, args.rounds + 1):
        round_start = ns.time
        for child in children:
            payload = f'cw{round_index:04d}n{child.node_id:02d}'
            ns.node_cmd(child.node_id, f'udp send {destination} {args.udp_port} {payload}')

        round_end = round_start + args.collection_window
        rounds.append(RoundWindow(round_index=round_index, start_s=round_start, end_s=round_end))

        seen_children = set()
        while ns.time < round_end - 1e-6:
            lines = ns.go(min(args.log_poll_interval, round_end - ns.time))
            now = ns.time
            for line in lines:
                parsed = parse_observation_token(line)
                if parsed is None:
                    continue
                observed_round, child_id = parsed
                if observed_round != round_index or child_id in seen_children:
                    continue
                child = child_by_id.get(child_id)
                if child is None:
                    continue
                observations.append(
                    Observation(
                        round_index=round_index,
                        child_id=child.node_id,
                        child_label=child.label,
                        extaddr=child.extaddr,
                        packet_number=-1,
                        sniff_time_s=now,
                        delay_from_round_start_ms=(now - round_start) * 1000.0,
                        phase_offset_ms=((now % args.csl_period_s) / args.csl_period_s) * args.csl_period_ms,
                        csl_phase_units=-1,
                        csl_phase_ie_ms=math.nan,
                        csl_period_units=-1,
                        frame_kind='udp_rx_log',
                    )
                )
                seen_children.add(child_id)

        elapsed = ns.time - round_start
        if elapsed < args.round_interval - 1e-6:
            ns.go(args.round_interval - elapsed)
    observations.sort(key=lambda row: (row.round_index, row.child_id))
    return rounds, observations


def circular_mean_ms(samples_ms: Sequence[float], period_ms: float) -> Tuple[float, float]:
    if not samples_ms:
        return math.nan, math.nan

    angles = [2.0 * math.pi * sample / period_ms for sample in samples_ms]
    sin_sum = sum(math.sin(angle) for angle in angles)
    cos_sum = sum(math.cos(angle) for angle in angles)
    mean_angle = math.atan2(sin_sum, cos_sum)
    if mean_angle < 0:
        mean_angle += 2.0 * math.pi
    mean_phase_ms = mean_angle * period_ms / (2.0 * math.pi)

    radius = math.hypot(sin_sum, cos_sum) / len(samples_ms)
    if radius <= 0.0:
        circular_std_ms = period_ms / math.sqrt(12.0)
    elif radius >= 1.0:
        circular_std_ms = 0.0
    else:
        circular_std_ms = math.sqrt(max(0.0, -2.0 * math.log(radius))) * period_ms / (2.0 * math.pi)
    return mean_phase_ms, circular_std_ms


def build_child_summaries(children: Sequence[ChildNode], observations: Sequence[Observation],
                          total_rounds: int, period_ms: float) -> List[ChildSummary]:
    by_child: Dict[int, List[Observation]] = {child.node_id: [] for child in children}
    for observation in observations:
        by_child[observation.child_id].append(observation)

    summaries: List[ChildSummary] = []
    for child in children:
        child_observations = by_child[child.node_id]
        phases = [row.phase_offset_ms for row in child_observations]
        mean_phase_ms, circular_std_ms = circular_mean_ms(phases, period_ms)
        summaries.append(
            ChildSummary(
                child_id=child.node_id,
                child_label=child.label,
                extaddr=child.extaddr,
                samples=len(child_observations),
                coverage=(len(child_observations) / total_rounds) if total_rounds else 0.0,
                mean_phase_ms=mean_phase_ms,
                circular_std_ms=circular_std_ms,
                min_phase_ms=min(phases) if phases else math.nan,
                max_phase_ms=max(phases) if phases else math.nan,
            )
        )

    summaries.sort(key=lambda row: row.child_id)
    return summaries


def build_wake_gaps(observations: Sequence[Observation], period_ms: float) -> List[WakeGap]:
    by_round: Dict[int, List[Observation]] = {}
    for observation in observations:
        by_round.setdefault(observation.round_index, []).append(observation)

    wake_gaps: List[WakeGap] = []
    for round_index, round_observations in sorted(by_round.items()):
        if len(round_observations) < 2:
            continue
        ordered = sorted(round_observations, key=lambda row: row.phase_offset_ms)
        for index, current in enumerate(ordered):
            nxt = ordered[(index + 1) % len(ordered)]
            if index + 1 < len(ordered):
                gap_ms = nxt.phase_offset_ms - current.phase_offset_ms
            else:
                gap_ms = period_ms - current.phase_offset_ms + ordered[0].phase_offset_ms
            wake_gaps.append(
                WakeGap(
                    round_index=round_index,
                    from_child_id=current.child_id,
                    from_child_label=current.child_label,
                    to_child_id=nxt.child_id,
                    to_child_label=nxt.child_label,
                    gap_ms=gap_ms,
                )
            )

    return wake_gaps


def build_sample_intervals(children: Sequence[ChildNode], observations: Sequence[Observation]) -> List[SampleInterval]:
    by_child: Dict[int, List[Observation]] = {child.node_id: [] for child in children}
    for observation in observations:
        by_child[observation.child_id].append(observation)

    intervals: List[SampleInterval] = []
    for child in children:
        child_observations = sorted(by_child[child.node_id], key=lambda row: row.sniff_time_s)
        for previous, current in zip(child_observations, child_observations[1:]):
            intervals.append(
                SampleInterval(
                    child_id=child.node_id,
                    child_label=child.label,
                    interval_ms=(current.sniff_time_s - previous.sniff_time_s) * 1000.0,
                )
            )
    return intervals


def percentile(values: Sequence[float], q: float) -> float:
    if not values:
        return math.nan
    ordered = sorted(values)
    if len(ordered) == 1:
        return ordered[0]
    position = (len(ordered) - 1) * q
    lower = math.floor(position)
    upper = math.ceil(position)
    if lower == upper:
        return ordered[lower]
    weight = position - lower
    return ordered[lower] * (1.0 - weight) + ordered[upper] * weight


def build_summary_metrics(args: argparse.Namespace, rounds: Sequence[RoundWindow], observations: Sequence[Observation],
                          summaries: Sequence[ChildSummary], wake_gaps: Sequence[WakeGap],
                          sample_intervals: Sequence[SampleInterval]) -> Dict[str, float]:
    expected_gap_ms = args.csl_period_ms / max(args.num_ssed, 1)
    gap_values = [row.gap_ms for row in wake_gaps]
    interval_values = [row.interval_ms for row in sample_intervals]
    crowded_threshold_ms = expected_gap_ms * 0.5
    return {
        'total_expected_observations': float(args.num_ssed * len(rounds)),
        'matched_observations': float(len(observations)),
        'mean_child_coverage': statistics.mean(summary.coverage for summary in summaries) if summaries else math.nan,
        'expected_gap_ms': expected_gap_ms,
        'minimum_observed_gap_ms': min(gap_values) if gap_values else math.nan,
        'wake_gap_p10_ms': percentile(gap_values, 0.10),
        'wake_gap_p50_ms': percentile(gap_values, 0.50),
        'wake_gap_p90_ms': percentile(gap_values, 0.90),
        'crowded_threshold_ms': crowded_threshold_ms,
        'crowded_fraction': (
            sum(1 for value in gap_values if value < crowded_threshold_ms) / len(gap_values)
            if gap_values else math.nan
        ),
        'sample_interval_p50_ms': percentile(interval_values, 0.50),
    }


def format_metric(value: float, suffix: str = '', digits: int = 3) -> str:
    if isinstance(value, float) and math.isnan(value):
        return 'n/a'
    return f'{value:.{digits}f}{suffix}'


def markdown_escape(value: str) -> str:
    return value.replace('|', '\\|')


def build_markdown_table(headers: Sequence[str], rows: Sequence[Sequence[str]]) -> str:
    lines = [
        '| ' + ' | '.join(headers) + ' |',
        '| ' + ' | '.join(['---'] * len(headers)) + ' |',
    ]
    for row in rows:
        lines.append('| ' + ' | '.join(markdown_escape(cell) for cell in row) + ' |')
    return '\n'.join(lines)


def build_round_phase_order_rows(observations: Sequence[Observation], limit: int = 16) -> List[Sequence[str]]:
    by_round: Dict[int, List[Observation]] = {}
    for observation in observations:
        by_round.setdefault(observation.round_index, []).append(observation)

    rows: List[Sequence[str]] = []
    for round_index in sorted(by_round.keys())[:limit]:
        ordered = sorted(by_round[round_index], key=lambda row: row.phase_offset_ms)
        order_text = ' -> '.join(
            f'{row.child_label} ({row.phase_offset_ms:.1f} ms)'
            for row in ordered
        )
        rows.append((str(round_index), order_text))
    return rows


def build_smallest_gap_rows(wake_gaps: Sequence[WakeGap], limit: int = 12) -> List[Sequence[str]]:
    rows: List[Sequence[str]] = []
    for gap in sorted(wake_gaps, key=lambda row: row.gap_ms)[:limit]:
        rows.append((
            str(gap.round_index),
            gap.from_child_label,
            gap.to_child_label,
            f'{gap.gap_ms:.3f}',
        ))
    return rows


def build_overview_markdown(args: argparse.Namespace, rounds: Sequence[RoundWindow], children: Sequence[ChildNode],
                            observations: Sequence[Observation], summaries: Sequence[ChildSummary],
                            wake_gaps: Sequence[WakeGap], sample_intervals: Sequence[SampleInterval],
                            paths: Dict[str, Path]) -> str:
    metrics = build_summary_metrics(args, rounds, observations, summaries, wake_gaps, sample_intervals)
    lines = [
        '# CSL Wake-up Crowding Overview',
        '',
        '이 파일은 실험 설정, 핵심 crowding 지표, child별 phase 분포 요약, wake gap 요약을 한 번에 보기 위한 종합본이다.',
        '',
        '## Experiment Setup',
        '',
        build_markdown_table(
            ('Item', 'Value'),
            (
                ('Parent router/leader', '1'),
                ('SSED children', str(args.num_ssed)),
                ('CSL period', f'{args.csl_period_ms:.3f} ms ({args.csl_period_us} us)'),
                ('CSL timeout', f'{args.csl_timeout} s'),
                ('Radiomodel', args.radiomodel),
                ('Observation rounds', str(len(rounds))),
                ('Round interval', f'{args.round_interval:.3f} s'),
                ('Collection window', f'{args.collection_window:.3f} s'),
                ('Parent radio range', str(args.radio_range)),
                ('Child ring radius', str(args.child_radius)),
            ),
        ),
        '',
        '## Key Metrics',
        '',
        build_markdown_table(
            ('Metric', 'Value'),
            (
                ('Matched observations', f'{int(metrics["matched_observations"])} / {int(metrics["total_expected_observations"])}'),
                ('Mean child coverage', format_metric(metrics['mean_child_coverage'])),
                ('Expected uniform wake gap', format_metric(metrics['expected_gap_ms'], ' ms')),
                ('Minimum observed wake gap', format_metric(metrics['minimum_observed_gap_ms'], ' ms')),
                ('Wake gap p10 / p50 / p90',
                 f'{format_metric(metrics["wake_gap_p10_ms"], " ms")} / {format_metric(metrics["wake_gap_p50_ms"], " ms")} / {format_metric(metrics["wake_gap_p90_ms"], " ms")}'),
                ('Crowded gap threshold', format_metric(metrics['crowded_threshold_ms'], ' ms')),
                ('Fraction below crowded threshold', format_metric(metrics['crowded_fraction'] * 100.0 if not math.isnan(metrics['crowded_fraction']) else math.nan, ' %')),
                ('Median sample interval', format_metric(metrics['sample_interval_p50_ms'], ' ms')),
            ),
        ),
        '',
        '## Graphs',
        '',
    ]

    if paths['plot'].exists():
        lines.extend([
            f'![CSL crowding plots]({paths["plot"].name})',
            '',
            '- 상단 그래프: child별 phase 분포',
            '- 하단 그래프: 동일 CSL period 안의 인접 wake-up gap histogram',
            '',
        ])
    else:
        lines.extend([
            '_Plot PNG was not generated. Rerun without `--skip-plot` to embed the graph here._',
            '',
        ])

    lines.extend([
        '## Per-Child Phase Summary',
        '',
        build_markdown_table(
            ('Child', 'Samples', 'Coverage', 'Mean Phase (ms)', 'Circular Std (ms)', 'Min (ms)', 'Max (ms)'),
            tuple(
                (
                    summary.child_label,
                    str(summary.samples),
                    f'{summary.coverage:.3f}',
                    format_metric(summary.mean_phase_ms),
                    format_metric(summary.circular_std_ms),
                    format_metric(summary.min_phase_ms),
                    format_metric(summary.max_phase_ms),
                )
                for summary in summaries
            ),
        ),
        '',
        '## Smallest Wake Gaps',
        '',
        build_markdown_table(
            ('Round', 'From', 'To', 'Gap (ms)'),
            build_smallest_gap_rows(wake_gaps),
        ) if wake_gaps else '_No wake-gap rows available._',
        '',
        '## Round-by-Round Phase Order',
        '',
        build_markdown_table(
            ('Round', 'Observed child order inside one CSL period'),
            build_round_phase_order_rows(observations),
        ) if observations else '_No observation rows available._',
        '',
        '## Artifacts',
        '',
        f'- [Report TXT]({paths["report"].name})',
        f'- [Observations CSV]({paths["observations"].name})',
        f'- [Child Summary CSV]({paths["child_summary"].name})',
        f'- [Wake Gaps CSV]({paths["wake_gaps"].name})',
        f'- [Sample Intervals CSV]({paths["sample_intervals"].name})',
        f'- [Rounds CSV]({paths["rounds"].name})',
        f'- [PCAP]({paths["pcap"].name})',
    ])
    if paths['plot'].exists():
        lines.append(f'- [Plot PNG]({paths["plot"].name})')

    return '\n'.join(lines)


def write_rounds_csv(path: Path, rounds: Sequence[RoundWindow]) -> None:
    with path.open('w', newline='') as handle:
        writer = csv.DictWriter(handle, fieldnames=['round', 'start_s', 'end_s'])
        writer.writeheader()
        for round_window in rounds:
            writer.writerow({
                'round': round_window.round_index,
                'start_s': f'{round_window.start_s:.6f}',
                'end_s': f'{round_window.end_s:.6f}',
            })


def write_observations_csv(path: Path, observations: Sequence[Observation]) -> None:
    with path.open('w', newline='') as handle:
        writer = csv.DictWriter(handle, fieldnames=[
            'round', 'child_id', 'child_label', 'extaddr', 'packet_number', 'sniff_time_s',
            'delay_from_round_start_ms', 'phase_offset_ms', 'csl_phase_units', 'csl_phase_ie_ms',
            'csl_period_units', 'frame_kind',
        ])
        writer.writeheader()
        for observation in observations:
            writer.writerow({
                'round': observation.round_index,
                'child_id': observation.child_id,
                'child_label': observation.child_label,
                'extaddr': observation.extaddr,
                'packet_number': observation.packet_number,
                'sniff_time_s': f'{observation.sniff_time_s:.6f}',
                'delay_from_round_start_ms': f'{observation.delay_from_round_start_ms:.3f}',
                'phase_offset_ms': f'{observation.phase_offset_ms:.3f}',
                'csl_phase_units': observation.csl_phase_units,
                'csl_phase_ie_ms': f'{observation.csl_phase_ie_ms:.3f}',
                'csl_period_units': observation.csl_period_units,
                'frame_kind': observation.frame_kind,
            })


def write_child_summary_csv(path: Path, summaries: Sequence[ChildSummary]) -> None:
    with path.open('w', newline='') as handle:
        writer = csv.DictWriter(handle, fieldnames=[
            'child_id', 'child_label', 'extaddr', 'samples', 'coverage', 'mean_phase_ms',
            'circular_std_ms', 'min_phase_ms', 'max_phase_ms',
        ])
        writer.writeheader()
        for summary in summaries:
            writer.writerow({
                'child_id': summary.child_id,
                'child_label': summary.child_label,
                'extaddr': summary.extaddr,
                'samples': summary.samples,
                'coverage': f'{summary.coverage:.3f}',
                'mean_phase_ms': f'{summary.mean_phase_ms:.3f}',
                'circular_std_ms': f'{summary.circular_std_ms:.3f}',
                'min_phase_ms': f'{summary.min_phase_ms:.3f}',
                'max_phase_ms': f'{summary.max_phase_ms:.3f}',
            })


def write_wake_gaps_csv(path: Path, gaps: Sequence[WakeGap]) -> None:
    with path.open('w', newline='') as handle:
        writer = csv.DictWriter(handle, fieldnames=[
            'round', 'from_child_id', 'from_child_label', 'to_child_id', 'to_child_label', 'gap_ms',
        ])
        writer.writeheader()
        for gap in gaps:
            writer.writerow({
                'round': gap.round_index,
                'from_child_id': gap.from_child_id,
                'from_child_label': gap.from_child_label,
                'to_child_id': gap.to_child_id,
                'to_child_label': gap.to_child_label,
                'gap_ms': f'{gap.gap_ms:.3f}',
            })


def write_sample_intervals_csv(path: Path, intervals: Sequence[SampleInterval]) -> None:
    with path.open('w', newline='') as handle:
        writer = csv.DictWriter(handle, fieldnames=['child_id', 'child_label', 'interval_ms'])
        writer.writeheader()
        for interval in intervals:
            writer.writerow({
                'child_id': interval.child_id,
                'child_label': interval.child_label,
                'interval_ms': f'{interval.interval_ms:.3f}',
            })


def build_report_text(args: argparse.Namespace, rounds: Sequence[RoundWindow], children: Sequence[ChildNode],
                      observations: Sequence[Observation], summaries: Sequence[ChildSummary],
                      wake_gaps: Sequence[WakeGap], sample_intervals: Sequence[SampleInterval],
                      paths: Dict[str, Path]) -> str:
    metrics = build_summary_metrics(args, rounds, observations, summaries, wake_gaps, sample_intervals)

    lines = [
        'CSL Wake-up Crowding Report',
        '',
        'Topology',
        f'- Parent router/leader: 1',
        f'- SSED children: {args.num_ssed}',
        f'- CSL period: {args.csl_period_ms:.3f} ms ({args.csl_period_us} us)',
        f'- CSL timeout: {args.csl_timeout} s',
        f'- Radiomodel: {args.radiomodel}',
        f'- Low-traffic round interval: {args.round_interval:.3f} s',
        f'- Per-round collection window: {args.collection_window:.3f} s',
        f'- Observation rounds: {len(rounds)}',
        '',
        'Interpretation',
        '- Per-child phase distribution is estimated from the parent receive-log timestamp of each low-traffic child UDP trigger.',
        '- These receive times are used as a practical wake-up proxy for CSL children in this topology.',
        '- Wake-gap histogram uses adjacent child phase separation inside one CSL period for each round.',
        '- Smaller wake gaps imply stronger same-period crowding pressure on the parent channel schedule.',
        '',
        'Coverage',
        f'- Total matched observations: {len(observations)} / {args.num_ssed * len(rounds)} expected',
        f'- Mean child coverage: {statistics.mean(summary.coverage for summary in summaries):.3f}' if summaries else '- Mean child coverage: n/a',
        '',
        'Crowding Metrics',
        f'- Expected uniform wake gap: {metrics["expected_gap_ms"]:.3f} ms',
        f'- Minimum observed wake gap: {metrics["minimum_observed_gap_ms"]:.3f} ms' if not math.isnan(metrics['minimum_observed_gap_ms']) else '- Minimum observed wake gap: n/a',
        f'- Wake gap p10 / median / p90: {metrics["wake_gap_p10_ms"]:.3f} / {metrics["wake_gap_p50_ms"]:.3f} / {metrics["wake_gap_p90_ms"]:.3f} ms' if not math.isnan(metrics['wake_gap_p10_ms']) else '- Wake gap p10 / median / p90: n/a',
        f'- Fraction of wake gaps below 0.5 x uniform spacing ({metrics["crowded_threshold_ms"]:.3f} ms): {metrics["crowded_fraction"]:.3f}' if not math.isnan(metrics['crowded_fraction']) else '- Fraction of wake gaps below 0.5 x uniform spacing: n/a',
        f'- Observed sample interval median: {metrics["sample_interval_p50_ms"]:.3f} ms' if not math.isnan(metrics['sample_interval_p50_ms']) else '- Observed sample interval median: n/a',
        '',
        'Per-Child Phase Summary',
    ]

    for summary in summaries:
        lines.append(
            f'- {summary.child_label}: samples={summary.samples}, coverage={summary.coverage:.3f}, '
            f'mean_phase={summary.mean_phase_ms:.3f} ms, circ_std={summary.circular_std_ms:.3f} ms, '
            f'range=[{summary.min_phase_ms:.3f}, {summary.max_phase_ms:.3f}] ms'
        )

    lines.extend([
        '',
        'Artifacts',
        f'- Rounds CSV: {paths["rounds"]}',
        f'- Observations CSV: {paths["observations"]}',
        f'- Child summary CSV: {paths["child_summary"]}',
        f'- Wake gaps CSV: {paths["wake_gaps"]}',
        f'- Sample intervals CSV: {paths["sample_intervals"]}',
        f'- Overview Markdown: {paths["overview"]}',
        f'- PCAP: {paths["pcap"]}',
    ])
    if not args.skip_plot:
        lines.append(f'- Plot PNG: {paths["plot"]}')

    return '\n'.join(lines)


def write_report(path: Path, report_text: str) -> None:
    path.write_text(report_text)


def write_overview(path: Path, overview_text: str) -> None:
    path.write_text(overview_text)


def save_plot(path: Path, children: Sequence[ChildNode], observations: Sequence[Observation],
              summaries: Sequence[ChildSummary], wake_gaps: Sequence[WakeGap], args: argparse.Namespace) -> None:
    try:
        import matplotlib
        matplotlib.use('Agg')
        import matplotlib.pyplot as plt
    except Exception as exc:
        raise RuntimeError('matplotlib is required unless --skip-plot is used') from exc

    child_index = {child.node_id: index for index, child in enumerate(children)}
    summary_by_child = {summary.child_id: summary for summary in summaries}

    figure, axes = plt.subplots(2, 1, figsize=(14, 10), constrained_layout=True)

    for observation in observations:
        axes[0].scatter(
            observation.phase_offset_ms,
            child_index[observation.child_id],
            s=22,
            alpha=0.75,
            color='#2f6b7d',
        )

    for child in children:
        summary = summary_by_child.get(child.node_id)
        if summary is None or math.isnan(summary.mean_phase_ms):
            continue
        axes[0].scatter(summary.mean_phase_ms, child_index[child.node_id], s=80, color='#d14c32', marker='|')

    axes[0].set_title('Per-child phase distribution')
    axes[0].set_xlabel('Observed phase offset within CSL period (ms)')
    axes[0].set_ylabel('Child')
    axes[0].set_xlim(0.0, args.csl_period_ms)
    axes[0].set_yticks(range(len(children)))
    axes[0].set_yticklabels([child.label for child in children])
    axes[0].grid(True, axis='x', alpha=0.25)

    gap_values = [gap.gap_ms for gap in wake_gaps]
    if gap_values:
        bin_count = min(max(args.num_ssed * 2, 10), 64)
        axes[1].hist(gap_values, bins=bin_count, color='#bc6c25', alpha=0.85, edgecolor='black')
        axes[1].axvline(args.csl_period_ms / args.num_ssed, color='black', linestyle='--', linewidth=1.2,
                        label='uniform spacing')
        axes[1].legend()
    else:
        axes[1].text(0.5, 0.5, 'No wake-gap samples', ha='center', va='center', transform=axes[1].transAxes)

    axes[1].set_title('Wake-up gap histogram')
    axes[1].set_xlabel('Adjacent wake gap inside one CSL period (ms)')
    axes[1].set_ylabel('Count')
    axes[1].grid(True, axis='y', alpha=0.25)

    figure.savefig(path, dpi=140)
    plt.close(figure)


def run_case_study(args: argparse.Namespace, paths: Dict[str, Path]) -> str:
    ns: Optional[OTNS] = None
    rounds: List[RoundWindow] = []
    children: List[ChildNode] = []
    observations: List[Observation] = []

    try:
        ns = OTNS(sim_id=args.sim_id, otns_args=['-pcap', 'wpan-tap'])
        parent_id, child_ids = build_topology(ns, args)
        children = collect_child_nodes(ns, child_ids)
        destination = str(ns.get_mleid(parent_id))
        rounds, observations = run_rounds(ns, destination, children, args)
    finally:
        if ns is not None:
            try:
                ns.close()
            finally:
                ns.save_pcap(str(paths['results_dir']), paths['pcap'].name)

    summaries = build_child_summaries(children, observations, len(rounds), args.csl_period_ms)
    wake_gaps = build_wake_gaps(observations, args.csl_period_ms)
    sample_intervals = build_sample_intervals(children, observations)

    write_rounds_csv(paths['rounds'], rounds)
    write_observations_csv(paths['observations'], observations)
    write_child_summary_csv(paths['child_summary'], summaries)
    write_wake_gaps_csv(paths['wake_gaps'], wake_gaps)
    write_sample_intervals_csv(paths['sample_intervals'], sample_intervals)

    report_text = build_report_text(args, rounds, children, observations, summaries, wake_gaps, sample_intervals, paths)
    write_report(paths['report'], report_text)

    if not args.skip_plot:
        save_plot(paths['plot'], children, observations, summaries, wake_gaps, args)

    overview_text = build_overview_markdown(
        args,
        rounds,
        children,
        observations,
        summaries,
        wake_gaps,
        sample_intervals,
        paths,
    )
    write_overview(paths['overview'], overview_text)

    return report_text


def main() -> None:
    args = parse_args()
    paths = build_output_paths(args)
    report_text = run_case_study(args, paths)
    print(report_text)


if __name__ == '__main__':
    try:
        main()
    except OTNSExitedError as exc:
        if exc.exit_code != 0:
            raise