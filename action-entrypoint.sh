response=$(ecsctl $@)
echo "response=$(echo ecsctl $@)" >> $GITHUB_OUTPUT;
