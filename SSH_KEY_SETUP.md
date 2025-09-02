# Alternative: SSH Key Authentication Workflow

If you prefer to use SSH key authentication (recommended for security), you can replace the "Upload to dl.furiatona.dev" step in the workflow with this version:

```yaml
- name: Upload to dl.furiatona.dev (SSH Key)
  run: |
    VERSION="${GITHUB_REF#refs/tags/}"
    echo "Uploading version ${VERSION}"
    
    # Create version directory structure
    mkdir -p "uploads/${VERSION}"
    cp dist/* "uploads/${VERSION}/"
    
    # Setup SSH key
    mkdir -p ~/.ssh
    echo "${{ secrets.SSH_PRIVATE_KEY }}" > ~/.ssh/id_rsa
    chmod 600 ~/.ssh/id_rsa
    ssh-keyscan -H "${{ secrets.SSH_HOST }}" >> ~/.ssh/known_hosts
    
    # Upload via SSH with key authentication
    echo "Uploading to dl.furiatona.dev..."
    scp -r "uploads/${VERSION}/" "${{ secrets.SSH_USER }}@${{ secrets.SSH_HOST }}:/path/to/dl.furiatona.dev/azctl/"
    
    # Create/update latest symlink
    ssh "${{ secrets.SSH_USER }}@${{ secrets.SSH_HOST }}" "cd /path/to/dl.furiatona.dev/azctl && ln -sfn ${VERSION} latest"
    
    echo "✅ Upload complete! Files available at dl.furiatona.dev/azctl/${VERSION}/"
    echo "✅ Latest symlink updated to point to ${VERSION}"
```

## SSH Key Setup Instructions

1. **Generate SSH Key Pair:**
   ```bash
   ssh-keygen -t rsa -b 4096 -C "github-actions@furiatona.dev" -f ~/.ssh/github_deploy_key
   ```

2. **Add Public Key to Server:**
   ```bash
   # Copy the public key to your server
   cat ~/.ssh/github_deploy_key.pub | ssh user@dl.furiatona.dev "mkdir -p ~/.ssh && cat >> ~/.ssh/authorized_keys"
   ```

3. **Add Private Key to GitHub Secrets:**
   - Copy the entire content of `~/.ssh/github_deploy_key` (including the BEGIN and END lines)
   - Add it as `SSH_PRIVATE_KEY` secret in GitHub

4. **Test the Key:**
   ```bash
   ssh -i ~/.ssh/github_deploy_key user@dl.furiatona.dev "echo 'SSH key authentication successful'"
   ```
