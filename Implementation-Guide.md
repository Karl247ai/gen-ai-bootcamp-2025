Beginner-Friendly Implementation Guide for WSL
1. WSL Setup Check

# Check WSL version and directory
wsl --version
echo $HOME

# Create project in your WSL home directory
cd ~/Documents
mkdir -p lang-portal
cd lang-portal

# Set up Go in WSL. First, update your system. Then, Go Installation in WSL Ubuntu and Verify Go installation


sudo apt update && sudo apt upgrade -y

sudo apt install golang-go

go version     
# Using go version go1.18.1 linux/amd64

# Set up Go workspace in your WSL environment

# Add these to your ~/.bashrc
echo 'export GOPATH=$HOME/go' >> ~/.bashrc
echo 'export PATH=$PATH:/usr/local/go/bin:$GOPATH/bin' >> ~/.bashrc
source ~/.bashrc

# Now create your project

# Navigate to your project directory
cd ~/Documents/Genaibootcamp/lang-portal

# Initialize Go module
go mod init github.com/karl247ai/lang-portal

# Verify module creation
ls -la
cat go.mod

# Create project structure

# Create directories
mkdir -p cmd/server
mkdir -p internal/{api,models,repository,service}
mkdir -p migrations
mkdir -p config
mkdir -p test

# Verify and set up the development environment
# Verify Installation and Project Structure

# Check Go installation
go version

# Setting Up Development Environment
# First, let's install the tree command and verify our project structure:

# Install tree utility
sudo apt install tree

# Verify project structure again
tree .

# Here is the project structure

.
├── Documentation
├── Implementation-Guide.md
├── Implementation-Plan.md
├── cmd
│   └── server
├── config
├── go.mod
├── internal
│   ├── api
│   ├── models
│   ├── repository
│   └── service
├── migrations
├── reference_images
│   ├── Backend-Technical-Specs.md
│   └── Frontend-Technical-Specs.md
└── test

11 directories, 6 files

# Now, let's create our initial project files:

Create main Go file: cmd/server/main.go

Create VS Code workspace settings: .vscode/settings.json

Install required Go packages:

# Install Gin web framework
go get -u github.com/gin-gonic/gin

# Verify dependencies in go.mod
cat go.mod

Run the application:



Test the API: Open a new terminal and run: