#!/bin/bash
cd /home/debian
rm -rf Kagi-Session2API-MCP
git clone https://github.com/KSroido/Kagi-Session2API-MCP.git
cd Kagi-Session2API-MCP
echo "=== Clone complete ==="
echo "=== Directory structure ==="
find . -type f | head -50
echo "=== Git log ==="
git log --oneline -5
