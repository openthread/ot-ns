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

cmake_minimum_required(VERSION 3.10.2)
project(ot-rfsim VERSION 2.0.0)

set(CMAKE_CXX_EXTENSIONS OFF)
set(CMAKE_CXX_STANDARD 11)
set(CMAKE_C_STANDARD 99)

set(OT_PLATFORM "external" CACHE STRING "Target platform must be 'external'")
set(OT_PLATFORM_LIB "openthread-rfsim")
if(DEFINED ENV{OT_DIR})
    set(OT_DIR "$ENV{OT_DIR}")
    message( "-- [ot-rfsim] Caller has set OT_DIR to: ${OT_DIR}")
else()
    set(OT_DIR "${PROJECT_SOURCE_DIR}/../openthread")
endif()
if(DEFINED ENV{OTNS_NODE_TYPE})
    set(OTNS_NODE_TYPE "$ENV{OTNS_NODE_TYPE}")
    message( "-- [ot-rfsim] Caller has set OTNS_NODE_TYPE to: ${OTNS_NODE_TYPE}")
else()
    set(OTNS_NODE_TYPE "custom")
endif()
get_filename_component(OT_DIR "${OT_DIR}" REALPATH)
set(BUILD_DIR "build/ot-rfsim-${OTNS_NODE_TYPE}")
message( "-- [ot-rfsim] Using openthread stack for build       : ${OT_DIR}" )
message( "-- [ot-rfsim] Using openthread build output directory: ${BUILD_DIR}" )

set(CMAKE_ARCHIVE_OUTPUT_DIRECTORY ${PROJECT_BINARY_DIR}/lib)
set(CMAKE_LIBRARY_OUTPUT_DIRECTORY ${PROJECT_BINARY_DIR}/lib)
set(CMAKE_RUNTIME_OUTPUT_DIRECTORY ${PROJECT_BINARY_DIR}/bin)

add_definitions( -DRFSIM_SW_VERSION="${CMAKE_PROJECT_VERSION}" )

# an external openthread repo may be defined - its build files are stored in there, so that coverage info
# collection will work.
add_subdirectory(${OT_DIR} ${OT_DIR}/${BUILD_DIR})

target_compile_definitions(ot-config INTERFACE
    OPENTHREAD_CONFIG_FILE="openthread-core-rfsim-config.h"
    OPENTHREAD_PROJECT_CORE_CONFIG_FILE="openthread-core-rfsim-config.h"
    OPENTHREAD_CORE_CONFIG_PLATFORM_CHECK_FILE="openthread-core-rfsim-config-check.h"
)

target_include_directories(ot-config INTERFACE
    ${PROJECT_SOURCE_DIR}/src
)

add_subdirectory(src)
