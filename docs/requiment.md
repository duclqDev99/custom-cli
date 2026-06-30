Đây là cách mình thường làm cho các Tech Lead hoặc Senior Dev. Thay vì nhớ lệnh của 10 công cụ khác nhau, mình tạo một CLI riêng.

Ví dụ tên là:

dev

Sau đó mọi thứ đều đi qua nó.

dev init
dev sync
dev graph
dev memory
dev doctor
dev clean
Bước 1. Tạo thư mục
mkdir -p ~/bin
touch ~/bin/dev
chmod +x ~/bin/dev

Mở file

nano ~/bin/dev
Bước 2. Viết CLI

Ví dụ đơn giản:

#!/bin/bash

COMMAND=$1

case "$COMMAND" in

graph)
graphify extract .
;;

update)
graphify update .
;;

watch)
graphify watch .
;;

memory)
codebase-memory-mcp index
;;

doctor)
graphify --help >/dev/null

    if [ $? -eq 0 ]; then
        echo "✅ Graphify OK"
    else
        echo "❌ Graphify Missing"
    fi

    codebase-memory-mcp doctor
    ;;

\*)
echo ""
echo "Usage:"
echo " dev graph"
echo " dev update"
echo " dev watch"
echo " dev memory"
echo " dev doctor"
;;
esac
Bước 3. Thêm PATH

Mở

nano ~/.zshrc

thêm

export PATH="$HOME/bin:$PATH"

reload

source ~/.zshrc

Bây giờ ở bất cứ project nào

study-language/

chỉ cần

dev graph

thay vì

graphify extract .

Muốn update

dev update

Muốn check

dev doctor
Nhưng...

Đây mới chỉ là bản rất đơn giản.

Nếu là mình làm cho bạn

Mình sẽ viết khoảng 300~500 dòng Bash, chia thành nhiều module.

Ví dụ:

dev init

tự động

✓ Check Graphify

✓ Check MCP

✓ Check Git

✓ Check Node

✓ Check Python

✓ Check Docker

✓ Check Claude Code
dev sync

tự làm

graphify update .

↓

codebase-memory-mcp reindex

↓

git status

↓

Done
dev graph

tự

Nếu chưa có graph

↓

graphify extract .

↓

Nếu có

↓

graphify update .

Bạn không cần nhớ gì cả.

dev ai

sẽ kiểm tra

Claude

Gemini

Codex

OpenAI

Anthropic

Proxy

API Key

...
dev doctor

sẽ hiện

✓ Graphify

✓ MCP

✓ Docker

✓ Git

✓ Python

✓ Claude

✓ Node

✓ Redis

✓ PostgreSQL
Mình nghĩ còn có thể làm hay hơn nữa

Vì mình nhớ bạn đang dùng:

Claude Code
Graphify
Codebase Memory MCP
nhiều project Laravel/Node
Docker
n8n
Mac

Thay vì chỉ viết một script Bash, mình có thể giúp bạn xây dựng một CLI chuyên nghiệp bằng Node.js hoặc Go, cài bằng npm install -g hoặc brew, với các lệnh như:

dev init
dev sync
dev ai graph
dev ai memory
dev doctor
dev update
dev clean
dev stats

Nó sẽ có màu sắc, thanh tiến trình, kiểm tra lỗi, và rất dễ mở rộng khi sau này bạn thêm Graphiti, Semgrep, Codex CLI hay bất kỳ AI tool nào khác. Theo mình, đây là giải pháp đáng đầu tư vì bạn sẽ dùng nó mỗi ngày trong công việc.
