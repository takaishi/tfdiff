#!/bin/bash

for dir in ~/src/github.com/flyle-io/flyle-nexus-infra/terraform/env/staging/*/; do
    if [ -d "$dir" ]; then
        dirname=$(basename "$dir")
        if [ "$dirname" != "specific" ]; then
            echo "Directory: $dirname"
            ./dist/tfdiff compare ~/src/github.com/flyle-io/flyle-nexus-infra/terraform/env/{staging,production}/$dirname
        fi
    fi
done
