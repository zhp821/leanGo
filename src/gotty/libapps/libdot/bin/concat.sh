#!/bin/bash
# Copyright (c) 2012 The Chromium OS Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

FLAGS_HELP="
usage: ./bin/concat.sh -i FILE [-o FILE] [-f]

This script concatenates a list of files into a single file.  The list of files
is supplied via text file, using the -i|--input_from argument.  The filenames
can be relative to the working directory or located in a directory listed in
LIBDOT_SEARCH_PATH.  Absolute paths also work, but should be avoided when the
input file is intended to be used by others.

If the -f|--forever argument is provided, then the script will monitor the
list of input files and recreate the output when it detects a change.

The environment variable LIBDOT_SEARCH_PATH can be used to define a search
path.  It's consulted when resolving input files, and lets you specify relative
paths in the input file without imposing a parent directory structure.  Multiple
paths can be separated by the colon character.

There are a few directives that can be specified in the input file.  They are...

  @include FILE

  This can be used to include an additional list of files.  It's useful when
  you want to include a list of files specified by a separate project, or
  any time you want to compose lists of dependencies.

  The included FILE must be relative to the LIBDOT_SEARCH_PATH.

  If an included file specifies a file that is already part of the result it
  will not be duplicated.

  When an included file is being processed this script will change the current
  directory to the LIBDOT_SEARCH_PATH entry where the FILE was found.  This
  is to make certain that any scripts executed by an included @resource
  directive happen relative to a known location.

  @resource NAME TYPE SOURCE

    NAME - The resource NAME is that name that you'd use to fetch the resource
    with lib.resource.get(name)

    TYPE - If the resource type is 'raw' then the resource will be included
    without any munging.  Otherwise, the resource will be wrapped in a
    JavaScript string. If you specify the type as a valid mimetype then you'll
    be able to get the resource as a 'data:' url easily from
    lib.resource.getDataUrl(...).

    SOURCE - When specified as '< FILENAME' it is interpreted as a file from
    the LIBDOT_SEARCH_PATH search path.  When specified as '\$ shell command'
    then the output of the shell command is used as the value of the resource.

  This includes a resource in the output.  A resource can be a file on disk or
  the output of a shell command.  Resources are output as JavaScript strings by
  default but can also be the raw file contents or script output, which is
  useful when you want to include a JSON resource.

  The resource directive depends on libdot/js/lib_resource.js, but the
  dependency is not automatically injected.  It's up to your input file to
  include it.

  @echo DATA

  Echo's a static string to the output file.
"

source "$(dirname "$0")/common.sh"

DEFINE_boolean forever "$FLAGS_FALSE" \
  "Recreate the output file whenever one of the inputs changes. " f
DEFINE_string input_from '' \
  "A file containing the list of files to concatenate." i
DEFINE_string output '' \
  "The output file." o
DEFINE_boolean echo_inotify "$FLAGS_FALSE" \
  "Echo the inotify list to stdout." I

COMMAND_LINE="$(readlink -f $0) $@"
FLAGS "$@" || exit $?
eval set -- "${FLAGS_ARGV}"

NEWLINE=$'\n'

# Maximum file descriptor used.
CONCAT_MAX_FD=2

# List of files (absolute paths) we need to include in the inotify watch list.
# This may include files not in the output, like other concat files referenced
# with the @include directive.
CONCAT_INOTIFY_FILES=""

# List of files we've included in the output to be included in the header of
# the output.  These paths should be as specified in the concat source list
# so they're short and relative to the LIBDOT_SEARCH_PATH.
CONCAT_HEADER_FILES=""

# Output we've generated so far.
CONCAT_OUTPUT=""

# Echo the results to the FLAGS_output file or stdout if the flag is not
# provided.
function echo_results() {
  local saved_output="$CONCAT_OUTPUT"
  CONCAT_OUTPUT=""

  append_comment "This file was generated by libdot/bin/concat.sh."
  append_comment "It has been marked read-only for your safety.  Rather"
  append_comment "than edit it directly, please modify one of these source"
  append_comment "files..."
  append_comment

  for f in $CONCAT_HEADER_FILES; do
    append_comment "$f"
  done

  append_comment

  CONCAT_OUTPUT="$CONCAT_OUTPUT$NEWLINE$saved_output"

  if [ -z "$FLAGS_output" ]; then
    echo "${CONCAT_OUTPUT}"
  else
    rm -rf "$FLAGS_output"
    echo "${CONCAT_OUTPUT}" >> "$FLAGS_output"
    chmod a-w "$FLAGS_output"
  fi
}

# Append to the pending output.
#
# This also adds a trailing newline.
function append_output() {
  CONCAT_OUTPUT="${CONCAT_OUTPUT}$@$NEWLINE"
}

# Append a comment to the pending output.
#
# The output is wrapped to 79 columns, with each line (including the first)
# prefixed with "// ".
function append_comment() {
  local str=$*

  if [ -z "$str" ]; then
    append_output "//"
    return
  fi

  append_output "$(echo -n "${str}" | awk -v WIDTH=76 '
    {
      while (length>WIDTH) {
        print "// " substr($0,1,WIDTH);
        $0=substr($0,WIDTH+1);
      }

      print "// " $0;
    }'
  )"
}

# Append a JavaScript string to the pending output.
#
# The output is surrounded in single quote ("'") characters and wrapped to 79
# columns.  Wrapped lines are joined with a plus ("+").
#
# Single quotes found in the input are escaped.
function append_string() {
  local str=$*

  append_output "$(echo "${str//\'/\'}" | awk -v WIDTH=76 '
    {
      while (length>WIDTH) {
        print "\047" substr($0,1,WIDTH) "\047 +";
        $0=substr($0,WIDTH+1);
      }

      print "\047" $0 "\047 +";
    }

    END {
      print "\047\047";
    }'
  )"
}

# Convert data into a format that can be included in JavaScript and append it to
# the output.
#
# This makes the resource available via lib.resource.get(...), and depends
# on libdot/js/lib_resource.js.
#
# You can append the contents of a file or the output of a shell command.
#
# Resources are included in the list of files watched by inotify.
function append_resource() {
  local name="$1"
  local type="$2"
  local source="$3"

  local data

  if [ "${source:0:2}" == "\$ " ]; then
    # Resource generated by a command line.
    data=$(eval "${source:2}")
    insist

  elif [ "${source:0:2}" == "< " ]; then
    # Resource is the contents of an existing file.

    source="$(echo "${source:2}" | sed -e 's/[:space:]*//')"
    local abspath

    if [ "${source:0:1}" != '/' -a "${source:0:1}" != '.' ]; then
      abspath="$(search_file "${source}")"
      if [ -z "$abspath" ]; then
        echo_err "Can't find resource file: ${source}"
        return 1
      fi
    else
      abspath="$(readlink -f "$source")"
    fi

    CONCAT_HEADER_FILES="${CONCAT_HEADER_FILES} ${source}"
    CONCAT_INOTIFY_FILES="${CONCAT_INOTIFY_FILES}$NEWLINE$abspath"

    data="$(cat "$abspath")"

  else
    echo_err "Not sure how to interpret resource: $name $type $source"
    return 1
  fi

  append_output "lib.resource.add('$name', '$type',"

  if [ "$type" = "raw" ]; then
    # The resource should be the raw contents of the file or command output.
    # Great for json data.
    append_outout "${data}"

  else
    # Resource should be wrapped in a JS string.
    append_string "${data}"
  fi

  append_output ");"
  append_output
}

# Process a single line from a concat file.
function process_concat_line() {
  local line="$1"

  if [ "${line:0:1}" != "@" ]; then
    line="@file $line"
  fi

  echo_err "$line"

  if [ "${line:0:6}" = "@file " ]; then
    # Input line doesn't start with an "@", it's just a file to include
    # in the output.

    line="${line:6}"

    local abspath="$(search_file "$line")"
    if [ -z "$abspath" ]; then
      echo_err "File not found: $line"
      return 1
    fi

    if (echo -e "${CONCAT_INOTIFY_FILES}" | grep -xq "$abspath"); then
      echo_err "Skipping duplicate file."
      return 0
    fi

    CONCAT_HEADER_FILES="$CONCAT_HEADER_FILES $line"
    CONCAT_INOTIFY_FILES="$CONCAT_INOTIFY_FILES$NEWLINE$abspath"

    append_comment "SOURCE FILE: $line"
    append_output "$(cat "$abspath")"

  elif [ "${line:0:6}" = "@echo " ]; then
    append_output "${line:6}"

  elif [ "${line:0:6}" = "@eval " ]; then
    line=$(eval "${line:6}")
    insist
    append_output "$line"

  elif [ "${line:0:10}" = "@resource " ]; then
    local name="$(echo "${line:10}" | cut -d' ' -f1)"
    local type="$(echo "${line:10}" | cut -d' ' -f2)"
    local path="$(echo "${line:10}" | cut -d' ' -f3-)"
    insist append_resource "$name" "$type" "$path"

  else
    echo_err "Unknown directive: $line"
    return 1
  fi
}

# Process a concat file opened at a given file descriptor.
function process_concat_fd() {
  local fd=$1

  while read -r 0<&$fd READ_LINE; do
    local line=""

    READ_LINE=${READ_LINE##}  # Strip leading spaces.

    # Handle trailing escape as line continuation.
    while [ $(expr "$READ_LINE" : ".*\\\\$") != 0 ]; do
      READ_LINE=${READ_LINE%\\}  # Strip trailing escape.
      line="$line$READ_LINE"
      insist read -r 0<&$fd READ_LINE
    done

    line="$line$READ_LINE"

    if [ -z "$line" -o "${line:0:1}" = '#' ]; then
      # Skip blank lines and comments.
      continue
    fi

    if [ "${line:0:9}" = "@include " ]; then
      echo_err "$line"

      local relative_path="${line:9}"
      local absolute_path=$(search_file "$relative_path")

      insist process_concat_file "$absolute_path"
    else
      insist process_concat_line "$line"
    fi
  done
}

# Process a concat file specified by absolute path.
function process_concat_file {
  local absolute_path="$1"

  CONCAT_INOTIFY_FILES="$CONCAT_INOTIFY_FILES$NEWLINE$absolute_path"

  CONCAT_MAX_FD=$(($CONCAT_MAX_FD + 1))
  local fd=$CONCAT_MAX_FD

  eval "exec $fd< $absolute_path"
  insist

  pushd "$(dirname "$absolute_path")" > /dev/null
  insist process_concat_fd $fd
  popd > /dev/null

  eval "exec $fd>&-"
}

function main() {
  local input=""

  if [ -z "$FLAGS_input_from" ]; then
    echo_err "Missing argument: --input"
    exit 1
  fi

  if [ ! -z "$FLAGS_output" ]; then
    echo_err $(date "+%H:%M:%S") "- Re-creating: $FLAGS_output"
  fi

  insist process_concat_file "$FLAGS_input_from"

  echo_results

  if [ "$FLAGS_forever" = "$FLAGS_TRUE" ]; then
    echo -e '\a'

    inotifywait -qqe modify $0 $FLAGS_input_from $CONCAT_INOTIFY_FILES
    local err=$?
    if [[ $err != 0 && $err != 1 ]]; then
      echo_err "inotify exited with status code: $err"
      exit $err
    fi

    exec $COMMAND_LINE
  fi

  if [ "$FLAGS_echo_inotify" = "$FLAGS_TRUE" ]; then
    echo $FLAGS_input_from $CONCAT_INOTIFY_FILES
  fi

  return 0
}

main "$@"