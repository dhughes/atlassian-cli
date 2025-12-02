# Test Workflow for Jira Commands

This document provides a complete test workflow for the Jira commands.

## Test Workflow Steps

### 1. Create a Test Issue

```bash
./atl jira create-issue \
  --project FX \
  --type Task \
  --summary "CLI Test Issue - Delete Me" \
  --description "This is a test issue created by the atlassian CLI. Safe to delete."
```

**Expected output:**
```
✓ Created issue: FX-XXXX
  ID: 123456
  URL: https://yoursite.atlassian.net/rest/api/3/issue/123456

View issue: atl jira get-issue FX-XXXX
```

**→ Note the issue key (e.g., FX-XXXX) for the next steps!**

---

### 2. View the Created Issue

```bash
./atl jira get-issue FX-XXXX
```

**Expected output:**
Should show the issue with summary, description, status, etc.

---

### 3. Add a Comment

```bash
./atl jira add-comment FX-XXXX "This is a test comment from the CLI"
```

**Expected output:**
```
✓ Added comment to FX-XXXX
  Comment ID: 12345
```

**Verify:**
```bash
./atl jira get-issue FX-XXXX --json | jq '.fields.comment'
```

---

### 4. Edit the Issue

Update the summary:
```bash
./atl jira edit-issue FX-XXXX --summary "CLI Test Issue - UPDATED"
```

Update the description:
```bash
./atl jira edit-issue FX-XXXX --description "Updated description from CLI"
```

Update both:
```bash
./atl jira edit-issue FX-XXXX \
  --summary "CLI Test - Final Update" \
  --description "Final test description"
```

**Expected output:**
```
✓ Updated issue FX-XXXX
  Summary: CLI Test - Final Update
  Description: updated
```

**Verify:**
```bash
./atl jira get-issue FX-XXXX
```

---

### 5. Get Available Transitions

```bash
./atl jira get-transitions FX-XXXX
```

**Expected output:**
```
Available transitions for FX-XXXX:

  ID: 11    → Start Progress (to: In Progress)
  ID: 21    → Done (to: Done)
  ID: 31    → Won't Do (to: Won't Do)
  ...

To transition: atl jira transition-issue FX-XXXX <transition-id>
```

**→ Note the transition ID for "Won't Do" (e.g., 31)**

---

### 6. Transition to "Won't Do"

```bash
./atl jira transition-issue FX-XXXX 31
```

Replace `31` with the actual transition ID from step 5.

**Expected output:**
```
✓ Transitioned issue FX-XXXX

View updated issue: atl jira get-issue FX-XXXX
```

**Verify:**
```bash
./atl jira get-issue FX-XXXX
```

Should show status as "Won't Do" (or whatever your workflow calls it).

---

## Complete Test Script

Here's a complete script you can run (replace FX with your project key):

```bash
#!/bin/bash

PROJECT="FX"  # Change this to your project key

echo "=== Step 1: Creating test issue ==="
CREATE_OUTPUT=$(./atl jira create-issue \
  --project $PROJECT \
  --type Task \
  --summary "CLI Test Issue - Delete Me" \
  --description "Test issue from atlassian CLI")

ISSUE_KEY=$(echo "$CREATE_OUTPUT" | grep "Created issue:" | awk '{print $3}')
echo "Created: $ISSUE_KEY"
echo

echo "=== Step 2: Viewing issue ==="
./atl jira get-issue $ISSUE_KEY
echo

echo "=== Step 3: Adding comment ==="
./atl jira add-comment $ISSUE_KEY "Test comment from CLI"
echo

echo "=== Step 4: Editing issue ==="
./atl jira edit-issue $ISSUE_KEY \
  --summary "CLI Test - UPDATED" \
  --description "Updated via CLI"
echo

echo "=== Step 5: Getting transitions ==="
./atl jira get-transitions $ISSUE_KEY
echo

echo "=== Step 6: Transition to Won't Do ==="
echo "Find the transition ID from above and run:"
echo "./atl jira transition-issue $ISSUE_KEY <transition-id>"
```

---

## Additional Commands to Test

### Search for the test issue

```bash
./atl jira search-jql "project = FX AND summary ~ 'CLI Test'"
```

### View issue in JSON format

```bash
./atl jira get-issue FX-XXXX --json
```

### Get specific fields only

```bash
./atl jira get-issue FX-XXXX --fields summary,status,assignee
```

---

## Cleanup

After testing, you can delete the test issue through the Jira web UI or keep it as "Won't Do" for reference.
