#!/usr/bin/env bash
# List all metrics from K8s receivers as CSV for spreadsheet import.
# Parses metadata.yaml in each receiver directory. Run from the repository root.
#
# Usage:
#   ./list_k8s_receiver_metrics.sh [REPO_ROOT]
#   REPO_ROOT defaults to current directory.
#
# Output columns (CSV):
#   Component, Metric, Description, Type, Unit, Metric Type, Extra Comments, File

set -e

ROOT="${1:-.}"

# All K8s-related receiver directories. Only those with a metrics section will
# produce output; the rest are silently skipped.
RECEIVERS="k8sclusterreceiver kubeletstatsreceiver k8sobjectsreceiver k8slogreceiver k8seventsreceiver"

# CSV header
echo '"Component","Metric","Description","Type","Unit","Metric Type","Extra Comments","File"'

for receiver_dir in $RECEIVERS; do
  meta="$ROOT/receiver/$receiver_dir/metadata.yaml"
  [ -f "$meta" ] || continue

  # Extract the canonical component name from the "type:" field at the top.
  component=$(awk '/^type:/ { print $2; exit }' "$meta")

  awk -v component="$component" -v file="$meta" '
    function csv_escape(s) {
      gsub(/"/, "\"\"", s)
      return "\"" s "\""
    }

    function emit() {
      if (current == "") return

      # Build the extra-comments field from collected flags.
      extra = ""
      if (enabled == "false") {
        extra = "optional"
      }
      if (stab != "") {
        if (extra != "") extra = extra "; "
        extra = extra "stability: " stab
      }
      if (mono != "") {
        if (extra != "") extra = extra "; "
        extra = extra "monotonic: " mono
      }
      if (attrs != "") {
        if (extra != "") extra = extra "; "
        extra = extra "attributes: " attrs
      }
      if (comments != "") {
        if (extra != "") extra = extra "; "
        extra = extra comments
      }

      row = csv_escape(component) "," \
            csv_escape(current)   "," \
            csv_escape(desc)      "," \
            csv_escape(vtype)     "," \
            csv_escape(unit)      "," \
            csv_escape(mtype)     "," \
            csv_escape(extra)     "," \
            csv_escape(file)
      print row

      # Reset per-metric state.
      current = ""; desc = ""; unit = ""; mtype = ""; vtype = ""
      mono = ""; stab = ""; attrs = ""; enabled = ""; comments = ""
    }

    # ---- Enter / leave the metrics section ----
    /^metrics:/ { in_metrics = 1; next }
    in_metrics && /^[a-zA-Z]/ { emit(); in_metrics = 0 }

    # ---- New metric name (2-space indent) ----
    in_metrics && /^  [a-zA-Z0-9._]+:/ {
      emit()
      in_attrs = 0
      current = $1; gsub(/:$/, "", current)
    }

    # ---- Scalar fields (4-space indent) ----
    in_metrics && /^    enabled:/     { enabled = $2 }
    in_metrics && /^    description:/ {
      desc = $0; sub(/^    description: */, "", desc)
      gsub(/^"/, "", desc); gsub(/"$/, "", desc)
    }
    in_metrics && /^    unit:/ {
      unit = $0; sub(/^    unit: */, "", unit)
      gsub(/^"/, "", unit); gsub(/"$/, "", unit)
    }

    # ---- Metric instrument type (4-space indent) ----
    in_metrics && /^    gauge:/     { mtype = "gauge" }
    in_metrics && /^    sum:/       { mtype = "sum" }
    in_metrics && /^    histogram:/ { mtype = "histogram" }

    # ---- Sub-fields of the instrument type (6-space indent) ----
    in_metrics && /^      value_type:/ { vtype = $2 }
    in_metrics && /^      monotonic:/  { mono = $2 }

    # ---- Stability level (6-space indent under stability:) ----
    in_metrics && /^      level:/ { stab = $2 }

    # ---- Attributes list (4-space key, 6-space list items) ----
    in_metrics && /^    attributes:/ {
      if (/\[\]/) next        # attributes: [] — empty list
      in_attrs = 1; next
    }
    in_attrs && /^      - / {
      a = $0; sub(/^      - /, "", a)
      if (attrs == "") attrs = a; else attrs = attrs ", " a
    }
    in_attrs && !/^      - / && !/^[[:space:]]*$/ { in_attrs = 0 }

    # ---- YAML comments inside the metrics block (2-space indent) ----
    in_metrics && /^  #/ {
      c = $0; sub(/^  # */, "", c)
      if (comments == "") comments = c; else comments = comments " " c
    }

    END { if (in_metrics) emit() }
  ' "$meta"
done
