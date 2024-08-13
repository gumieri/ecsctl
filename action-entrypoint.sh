#!/usr/bin/env bash

response=$(ecsctl $@)
echo "response=$(echo ecsctl $@)" >> $GITHUB_OUTPUT;
