#!/bin/bash
# ==============================================================================
# SCRIPT: init.sh
# DESCRIPTION: Interactive initialization wizard with built-in sandbox rollback
#              capabilities and state mutation protections for templates.
# ==============================================================================

# Ensure strict failure handling
set -e

# Define operational paths
CONFIG_FILE=".github/config.txt"
ENTRYPOINTS_FILE=".github/entrypoints.txt"
WORKFLOW_FILE=".github/workflows/main.yaml"
README_TMPL="README.tmpl"
README_FINAL="README.md"
LICENSE_FILE="LICENSE"

# Hidden location inside .git to prevent Go compiler or linters pollution
BKP_DIR=".git/.init_bkp"

# ================================================================================
# [RUN] Pre-execution safety and environment validations...
# ================================================================================

if [ ! -d ".git" ]; then
    echo "[ x ] Failure: This workspace is not initialized as a Git repository."
    echo "[ERR] Onboarding aborted. Please clone or init git before running."
    exit 1
fi

# ------------------------------------------------------------------------------
# INTERACTIVE MODE: Check for Finalization Flag
# ------------------------------------------------------------------------------
if [ "${1}" == "--finalize" ]; then
    echo "================================================================================"
    echo "[RUN] Finalizing architecture initialization lifecycle..."
    echo "================================================================================"
    
    if [ ! -d "$BKP_DIR" ]; then
        echo "[ x ] Fault: No active template initialization session found."
        echo "[ERR] Verification failed. Project might already be finalized."
        exit 1
    fi
    
    # Capture current branch context safely
    CURRENT_BRANCH=$(git branch --show-current)
    
    echo "[ . ] Staging baseline framework configurations on ${CURRENT_BRANCH}..."
    git add .
    git commit -m "feat: initial commit placeholder" --quiet
    
    # --------------------------------------------------------------------------
    # Cloud Provisioning: Create and Push Badges Branch
    # --------------------------------------------------------------------------
    echo "[ . ] Verifying cloud asset boundaries for test automation coverage badges..."

    # cleam remote cache before check
    git fetch --prune origin &> /dev/null

    # Check if the badges branch already exists locally or remotely
    if ! git rev-parse --verify origin/badges &> /dev/null && ! git rev-parse --verify badges &> /dev/null; then
        echo "================================================================================"
        echo "[ > ] ATTENTION: Provisioning remote 'badges' branch on GitHub."
        echo "      You might be prompted for your Git credentials below."
        echo "================================================================================"
        
        # Create an orphaned branch, force dumping index cache to ensure true emptiness
        git checkout --orphan badges --quiet
        
        # Create a minimal empty commit to establish branch presence
        git commit --allow-empty -m "chore: initialize empty badge storage partition" --quiet
        
        # Push the new branch to remote server upstream gateway
        git push origin badges || true
        
        # Return safely back to the original development branch workspace
        git checkout "$CURRENT_BRANCH" --quiet
        echo "[ v ] Telemetry branch generated and synchronized with cloud origin successfully."
    else
        echo "[ v ] Badge tracking partition presence already verified on environment channels."
    fi
    
    echo ""
    echo "[ . ] Purging local workspace blueprint backup files..."
    rm -rf "$BKP_DIR"
    rm -f "$README_TMPL"
    
    # Clear the initialization script file safely from native hard drive context
    SCRIPT_NAME=$(basename "$0")
    if [ -f "$SCRIPT_NAME" ]; then
        rm -f "$SCRIPT_NAME"
    fi
    
    echo "[ . ] Forging initial genesis structural repository commit milestone..."
    # Track mutations (deletions of bkp, readme.tmpl and init.sh)
    git add .
    
    # Refactor the temporary placeholder commit into the official clean history entry
    git commit --amend -m "feat: initialize template core files, metadata forms, and interactive setup wizard with rollback capabilities" --quiet
    
    echo "================================================================================"
    echo "[OKK] Application infrastructure secured and finalized flawlessly!"
    echo "================================================================================"
    exit 0
fi




# ================================================================================
# [RUN] Evaluating active sandbox state and processing rollbacks...
# ================================================================================

if [ -d "$BKP_DIR" ]; then
    echo "================================================================================"
    echo "[ > ] WARNING: Active installation detected!"
    echo "      Running this wizard again will DESTROY all previous configurations."
    echo "================================================================================"
    echo "[ > ] Do you want to reset your workspace and try again? (y/N):"
    read -r RESET_CONFIRM
    RESET_CONFIRM=$(echo "$RESET_CONFIRM" | tr '[:upper:]' '[:lower:]')
    
    if [ "$RESET_CONFIRM" != "y" ]; then
        echo "[ x ] Initialization process aborted by user decision."
        echo "[END] Execution halted safely."
        exit 0
    fi
    
    echo "[ . ] Cleaning up previously generated workspace footprints..."
    # CRITICAL: Clean the hidden submodule layout pointers inside Git cache database
    if git config --get "submodule..dev.url" &> /dev/null; then
        git submodule deinit -f .dev &> /dev/null || true
    fi
    
    # Force purge generated physical files and the internal Git modules cache folder
    rm -rf .dev
    rm -rf .git/modules/.dev
    rm -rf cmd
    rm -f go.mod go.sum "$README_FINAL"
    
    echo "[ . ] Restoring raw architecture assets from database backup..."
    cp -r "$BKP_DIR/.github" .
    cp "$BKP_DIR/$README_TMPL" .
    cp "$BKP_DIR/$LICENSE_FILE" .
else
    # First execution ever: Create the clean state backup partition
    mkdir -p "$BKP_DIR"
    cp -r .github "$BKP_DIR/"
    cp "$README_TMPL" "$BKP_DIR/"
    cp "$LICENSE_FILE" "$BKP_DIR/"
    echo "[ v ] Master environment backup state successfully locked."
fi

# Verification Gate: Ensure required files are present to execute the mutation
if [ ! -f "$CONFIG_FILE" ] || [ ! -f "$ENTRYPOINTS_FILE" ] || [ ! -f "$README_TMPL" ] || [ ! -f "$LICENSE_FILE" ]; then
    echo "[ x ] Failure: Missing critical core configuration or document components."
    echo "[ERR] Onboarding aborted due to unrecoverable file workspace drift."
    exit 1
fi




# ================================================================================
# [RUN] Initializing Core Project Metadata Questionnaire...
# ================================================================================

if command -v git &> /dev/null && git remote get-url origin &> /dev/null; then
    RAW_URL=$(git remote get-url origin)
    CLEAN_URL=$(echo "$RAW_URL" | sed -E 's|(https?://)?(git@)?||' | sed 's|:|/|' | sed 's/\.git$//')
    DEFAULT_MODULE="$CLEAN_URL"
else
    echo "[ x ] Failure: Could not locate remote origin URL configuration parameters."
    echo "[ERR] Onboarding aborted. Verify your local git environment stability state."
    exit 1
fi

echo "[ > ] Enter the Go module path [Default: $DEFAULT_MODULE]:"
read -r INPUT_MODULE
GO_MODULE="${INPUT_MODULE:-$DEFAULT_MODULE}"

echo ""
echo "[ > ] Choose the project type (app | cli | api | lib) [Default: app]:"
read -r INPUT_TYPE
CHOSEN_TYPE="${INPUT_TYPE:-app}"
CHOSEN_TYPE=$(echo "$CHOSEN_TYPE" | tr '[:upper:]' '[:lower:]')

if [ "$CHOSEN_TYPE" != "app" ] && [ "$CHOSEN_TYPE" != "cli" ] && [ "$CHOSEN_TYPE" != "api" ] && [ "$CHOSEN_TYPE" != "lib" ]; then
    echo "[ x ] Validation Failure: Invalid architectural type variant selection."
    echo "[ERR] Setup aborted. Allowed options are strictly: app, cli, api, lib."
    exit 1
fi

echo ""
echo "[ > ] Enter the main package name (will be forced to lowercase):"
read -r INPUT_MAIN_PKG
if [ -z "$INPUT_MAIN_PKG" ]; then
    echo "[ x ] Validation Failure: Main package identity parameter cannot be empty."
    echo "[ERR] Setup aborted due to unassigned operational context boundaries."
    exit 1
fi
MAIN_PKG_LOWER=$(echo "$INPUT_MAIN_PKG" | tr '[:upper:]' '[:lower:]')

echo ""
echo "[ > ] Enter the code coverage threshold percentage [Default: 80]:"
read -r INPUT_COVERAGE
COVERAGE_GATE="${INPUT_COVERAGE:-80}"

echo ""
echo "[ > ] Enter author's full name (for LICENSE file):"
read -r INPUT_FULL_NAME

echo ""
echo "[ > ] Enter corporate publisher / enterprise name:"
read -r INPUT_PUBLISHER

echo ""
echo "[ > ] Enter organization website URL:"
read -r INPUT_SITE

echo ""
echo "[ > ] Enter developer contact email address:"
read -r INPUT_EMAIL

echo ""
echo "[ > ] Enter a small punchy project description sentence:"
read -r INPUT_DESC




# ================================================================================
# [RUN] Injecting metadata constraints into target file stores...
# ================================================================================

echo "[ . ] Hydrating continuous integration database descriptors..."
sed -i "s|<project_type>|${CHOSEN_TYPE}|g" "$CONFIG_FILE"
sed -i "s|<main_pkg>|${MAIN_PKG_LOWER}|g" "$CONFIG_FILE"
sed -i "s|COVERAGE_THRESHOLD=80|COVERAGE_THRESHOLD=${COVERAGE_GATE}|g" "$CONFIG_FILE"

if [ "$CHOSEN_TYPE" != "lib" ]; then
    CMD_PATH="./cmd/${MAIN_PKG_LOWER}"
    echo "[ . ] Creating execution binary anchor directory structure at $CMD_PATH"
    mkdir -p "$CMD_PATH"
    
    cat << EOF > "${CMD_PATH}/main.go"
package main

import "fmt"

func main() {
	fmt.Println("Application execution interface initialized.")
}
EOF

    echo "$CMD_PATH" >> "$ENTRYPOINTS_FILE"
    echo "[ v ] Binary application compilation pathway activated inside data sheet."
else
    echo "[ v ] Pure reusable code framework library mode selected. Entrypoint compilation skipped."
fi




# ================================================================================
# [RUN] Personalizing legal license and human documentation files...
# ================================================================================

echo "[ . ] Mutating legal LICENSE framework..."
CURRENT_YEAR=$(date +"%Y")
sed -i "s|<year>|${CURRENT_YEAR}|g" "$LICENSE_FILE"
sed -i "s|<full_name>|${INPUT_FULL_NAME}|g" "$LICENSE_FILE"

echo "[ . ] Generating production documentation from template..."
cp "$README_TMPL" "$README_FINAL"

# Clean "://github.com" prefix from GO_MODULE to extract strictly the "Owner/Repo" pattern
REPO_PATH_CLEAN=$(echo "$GO_MODULE" | sed 's|^://github.com||')

REPO_NAME_RAW=$(basename "$GO_MODULE")
sed -i "s|<repo_name>|${REPO_NAME_RAW}|g" "$README_FINAL"
sed -i "s|<repo_path>|${REPO_PATH_CLEAN}|g" "$README_FINAL" # << ATUALIZADO COM A VARIÁVEL LIMPA
sed -i "s|<publisher>|${INPUT_PUBLISHER}|g" "$README_FINAL"
sed -i "s|<site>|${INPUT_SITE}|g" "$README_FINAL"
sed -i "s|<email>|${INPUT_EMAIL}|g" "$README_FINAL"
sed -i "s|<small_description>|${INPUT_DESC}|g" "$README_FINAL"




# ================================================================================
# [RUN] Attaching centralized governance development dependencies...
# ================================================================================

echo "[ . ] Mounting centralized governance submodule architecture into workspace..."
# Route submodules cleanly
git submodule add https://github.com/AeonDigital/Go-Core-Template-Dev.git .dev

echo "[ . ] Activating local Git execution gate lifecycle hook mappings..."
git config core.hooksPath .dev/hooks

echo "[ . ] Stabilizing platform multi-OS canonical line ending configurations..."
git config core.autocrlf false




# ================================================================================
# [RUN] Triggering native Go toolchain compilation setup pipelines...
# ================================================================================

echo "[ . ] Initializing production Go runtime layout boundaries..."
go mod init "$GO_MODULE"

if [ -f "go.mod" ]; then
    echo "[ v ] Module manifest created successfully. Pruning structural dependency trees..."
    go mod tidy &> /dev/null || true
else
    echo "[ x ] Technical Fault: Go compiler toolchain failed to initialize module."
    echo "[ERR] Module compilation generation collapsed."
    exit 1
fi




# ================================================================================
# [END] Wizard iteration complete
# ================================================================================

echo "================================================================================"
echo "[OKK] Generation successful! Test your project workspace components."
echo "      If you are SATISFIED, close the playground by executing:"
echo "      > ./init.sh --finalize"
echo "================================================================================"
