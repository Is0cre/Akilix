#!/bin/sh
set -eu

package=${1:-build/rpm/RPMS/x86_64/pensuse-0.0.1-0.m0.x86_64.rpm}
test -r "$package" || {
    echo "RPM not found: $package" >&2
    exit 1
}

files=$(rpm -qpl "$package")
for expected in \
    /usr/bin/pensuse \
    /usr/share/zsh/site-functions/_pensuse \
    /usr/share/bash-completion/completions/pensuse
do
    printf '%s\n' "$files" | grep -Fx "$expected" >/dev/null || {
        echo "RPM is missing $expected" >&2
        exit 1
    }
done

printf 'RPM payload verified: %s\n' "$package"
