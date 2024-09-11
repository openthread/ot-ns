#!/usr/bin/env python3
# Copyright (c) 2023, The OTNS Authors.
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

# node_clicker.py - a simple tool to convert mouse-clicked coordinates on a map image
# to a list of OTNS script lines for an OTNS Python script. These lines will add nodes
# at the clicked positions.

# required package python3-opencv (must be manually installed)
# (sudo apt-get install python3-opencv)

import cv2
import sys


def click_event(event, x, y, flags, params):

    if event == cv2.EVENT_LBUTTONDOWN or event == cv2.EVENT_RBUTTONDOWN:
        node_type = 'node_tp1'
        if event == cv2.EVENT_RBUTTONDOWN:
            node_type = 'node_tp2'
        print(f'ns.add({node_type}, x={x}, y={y}, radio_range={node_type}_radiorange)')
        cv2.circle(img, (x, y), 3, (255, 0, 0), -1)
        cv2.imshow('image', img)


if __name__ == "__main__":

    print('node_clicker.py - click pixel coordinates in the image, to generate OTNS node placement code.\n')
    if len(sys.argv) < 2:
        print('Usage: node_clicker.py <image-filename>')
        exit(1)

    imgFile = sys.argv[1]
    img = cv2.imread(imgFile, 1)

    cv2.imshow('image', img)
    cv2.setMouseCallback('image', click_event)

    print('# Copy/paste below code into your OTNS Python script')
    print('node_tp1 = \'router\'\nnode_tp2 = \'sed\'\nnode_tp1_radiorange = 160\nnode_tp2_radiorange = 160')

    cv2.waitKey(0)  # exit on keypress
    cv2.destroyAllWindows()
