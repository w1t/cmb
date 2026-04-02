#!/bin/bash
set -e

REPOS_DIR="test-repos"
mkdir -p "$REPOS_DIR"

echo "Setting up test repositories..."
echo ""

# Format: "URL|COMMIT|NAME|DESCRIPTION"
repos=(
  "https://github.com/go-chi/chi.git|v5.0.8|chi|Lightweight HTTP router"
  "https://github.com/gorilla/mux.git|v1.8.1|mux|URL router and dispatcher"
  "https://github.com/sirupsen/logrus.git|v1.9.3|logrus|Structured logger"
  "https://github.com/uber-go/zap.git|v1.26.0|zap|Fast structured logger"
  "https://github.com/spf13/cobra.git|v1.8.0|cobra|CLI framework"
)

for entry in "${repos[@]}"; do
  IFS='|' read -r url commit name description <<< "$entry"

  target="$REPOS_DIR/$name"

  if [ -d "$target/.git" ]; then
    echo "✓ $name ($description) - already exists"
    continue
  fi

  echo "→ Cloning $name ($description) at $commit..."
  git clone --quiet "$url" "$target"

  cd "$target"
  git checkout --quiet "$commit"
  cd - > /dev/null

  echo "✓ $name ready"
done

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "All test repositories ready!"
echo ""
echo "Available repositories:"
for entry in "${repos[@]}"; do
  IFS='|' read -r url commit name description <<< "$entry"
  printf "  • %-10s %s\n" "$name" "$description"
done
echo ""
echo "You can now run tasks with:"
echo "  ./cmb run --agent opencode --task task/your-task.yaml"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
