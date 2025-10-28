# Goo CLI - TODO and Future Improvements

## Current Status

✅ **Completed:**
- Full CLI structure with Cobra framework
- All major command groups implemented:
  - Config management
  - Request operations
  - Propose operations
  - Dispute operations
  - Vote operations (commit-reveal)
  - Query operations
  - Admin operations
- Print mode for transaction visualization
- Vote data local storage
- Configuration management
- Utility functions (hash, format, time)
- Complete documentation (README, EXAMPLES, ARCHITECTURE)
- Build scripts and Makefile
- All commands tested and working

## Phase 1: Core Functionality (Next Steps)

### 1.1 Implement Real Transaction Execution

- [ ] Replace print mode with actual gnokey execution in `executor.go`
- [ ] Parse and handle transaction results
- [ ] Error handling for failed transactions
- [ ] Transaction confirmation and status checking
- [ ] Gas estimation and adjustment

**Priority:** HIGH
**Effort:** Medium

### 1.2 Vote Data Management

- [ ] Implement actual file I/O for vote data
  - [ ] Save vote data to `~/.goo/votes/<id>.json`
  - [ ] Load vote data from files
  - [ ] Validate vote data integrity
- [ ] Add vote data backup/restore commands
- [ ] Add vote data export/import
- [ ] List all saved votes

**Priority:** HIGH
**Effort:** Low

### 1.3 Enhanced Error Handling

- [ ] Better error messages for common issues
- [ ] Transaction validation before submission
- [ ] Check balances before transactions
- [ ] Verify request exists before operations
- [ ] Check timing constraints (deadlines, periods)

**Priority:** HIGH
**Effort:** Medium
