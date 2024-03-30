#
#  Copyright (c) 2022-2024, The OpenThread Authors.
#  All rights reserved.
#
#  Redistribution and use in source and binary forms, with or without
#  modification, are permitted provided that the following conditions are met:
#  1. Redistributions of source code must retain the above copyright
#     notice, this list of conditions and the following disclaimer.
#  2. Redistributions in binary form must reproduce the above copyright
#     notice, this list of conditions and the following disclaimer in the
#     documentation and/or other materials provided with the distribution.
#  3. Neither the name of the copyright holder nor the
#     names of its contributors may be used to endorse or promote products
#     derived from this software without specific prior written permission.
#
#  THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS"
#  AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
#  IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE
#  ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE
#  LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR
#  CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF
#  SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS
#  INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN
#  CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE)
#  ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE
#  POSSIBILITY OF SUCH DAMAGE.
#

add_library(openthread-rfsim-config INTERFACE)

add_library(openthread-rfsim
    alarm.c
    ble.c
    crypto.c
    diag.c
    dso_transport.c
    entropy.c
    event-sim.c
    flash.c
    infra_if.c
    logging.c
    misc.c
    platform-rfsim.c
    radio.c
    system.c
    trel.c
    uart.c
    $<TARGET_OBJECTS:openthread-platform-utils>
)

find_library(LIBRT rt)
if(LIBRT)
    target_link_libraries(openthread-rfsim PRIVATE ${LIBRT})
endif()

target_link_libraries(openthread-rfsim PRIVATE
    openthread-platform
    openthread-rfsim-config
    ot-config
)

target_compile_options(openthread-rfsim PRIVATE
    ${OT_CFLAGS}
)

target_include_directories(openthread-rfsim PRIVATE
    ${OT_PUBLIC_INCLUDES}
    ${OT_DIR}/examples/platforms
    ${PROJECT_SOURCE_DIR}/src
)
