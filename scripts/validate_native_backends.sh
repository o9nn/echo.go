#!/usr/bin/env bash
# Validate Echo native backend boundaries before running cgo-heavy tests.
#
# This script deliberately separates three concerns:
#   1. the maintained source-based llama.cpp binding at ./llama;
#   2. the GGML backend package at ./ml/backend/ggml; and
#   3. the legacy direct-link wrapper at ./core/inference/llama, which requires
#      prebuilt libs/libllama and libs/libggml* artifacts plus -tags llama_legacy.
#
# Usage:
#   scripts/validate_native_backends.sh
#   scripts/validate_native_backends.sh --strict-legacy
#   scripts/validate_native_backends.sh --build-legacy-libs

set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$repo_root"

strict_legacy=0
build_legacy=0
for arg in "$@"; do
  case "$arg" in
    --strict-legacy) strict_legacy=1 ;;
    --build-legacy-libs) build_legacy=1 ;;
    -h|--help)
      sed -n '1,24p' "$0"
      exit 0
      ;;
    *)
      echo "unknown argument: $arg" >&2
      exit 2
      ;;
  esac
done

need() {
  if ! command -v "$1" >/dev/null 2>&1; then
    echo "missing required command: $1" >&2
    return 1
  fi
}

check_file() {
  local path="$1"
  if [ -f "$path" ]; then
    echo "ok: $path"
  else
    echo "missing: $path"
    return 1
  fi
}

status=0

echo "== cgo toolchain =="
need go || status=1
if ! command -v cc >/dev/null 2>&1 && ! command -v gcc >/dev/null 2>&1 && ! command -v clang >/dev/null 2>&1; then
  echo "missing: C compiler (cc/gcc/clang)"
  status=1
else
  echo "ok: C compiler available"
fi

echo
echo "== maintained source-based llama binding =="
if CGO_ENABLED=1 go test ./llama -run '^$' -count=1; then
  echo "ok: ./llama compiles against vendored llama.cpp sources"
else
  echo "failed: ./llama compile probe"
  status=1
fi

echo
echo "== GGML backend package =="
if CGO_ENABLED=1 go test ./ml/backend/ggml -run '^$' -count=1; then
  echo "ok: ./ml/backend/ggml compiles"
else
  echo "failed: ./ml/backend/ggml compile probe"
  status=1
fi

legacy_lib_dir="$repo_root/libs"
legacy_libs=(
  "$legacy_lib_dir/libllama.a"
  "$legacy_lib_dir/libggml.a"
  "$legacy_lib_dir/libggml-base.a"
  "$legacy_lib_dir/libggml-cpu.a"
)

echo
echo "== legacy core/inference/llama direct-link libraries =="
legacy_missing=0
for lib in "${legacy_libs[@]}"; do
  check_file "$lib" || legacy_missing=1
done

if [ "$build_legacy" -eq 1 ]; then
  echo
echo "== legacy library build attempt =="
  need cmake || status=1
  if command -v cmake >/dev/null 2>&1; then
    build_dir="$repo_root/build/legacy-native-libs"
    cmake -S "$repo_root/llama/llama.cpp" -B "$build_dir" -DBUILD_SHARED_LIBS=OFF -DLLAMA_CURL=OFF
    cmake --build "$build_dir" --target llama ggml ggml-base ggml-cpu -j"$(nproc)"
    mkdir -p "$legacy_lib_dir"
    find "$build_dir" -type f \( -name 'libllama.a' -o -name 'libggml*.a' \) -exec cp -f {} "$legacy_lib_dir" \;
    legacy_missing=0
    for lib in "${legacy_libs[@]}"; do
      check_file "$lib" || legacy_missing=1
    done
  fi
fi

if [ "$legacy_missing" -eq 0 ]; then
  echo "ok: legacy libraries present"
  if CGO_ENABLED=1 go test -tags llama_legacy ./core/inference/llama -run '^$' -count=1; then
    echo "ok: legacy core/inference/llama wrapper links"
  else
    echo "failed: legacy core/inference/llama wrapper link probe"
    status=1
  fi
else
  echo "legacy libraries unavailable; default package build remains supported by llama_unavailable.go"
  if [ "$strict_legacy" -eq 1 ]; then
    status=1
  fi
fi

exit "$status"
