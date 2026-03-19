#!/bin/bash
# Auto-commit after file changes
INPUT=$(cat)
TOOL_NAME=$(echo "$INPUT" | jq -r '.toolName')
RESULT_TYPE=$(echo "$INPUT" | jq -r '.toolResult.resultType')

# Only commit after successful file operations
if [ "$RESULT_TYPE" = "success" ]; then
  case "$TOOL_NAME" in
    edit|create)
      cd "$(echo "$INPUT" | jq -r '.cwd')"
      if [ -n "$(git status --porcelain 2>/dev/null)" ]; then
        git add -A
        git commit -m "Auto-save: $TOOL_NAME operation" --no-verify 2>/dev/null || true
      fi
      ;;
  esac
fi
