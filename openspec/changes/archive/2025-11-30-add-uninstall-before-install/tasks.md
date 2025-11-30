# Implementation Tasks

## Phase 1: Core Implementation

- [x] **Update cmd/install.go to call uninstall before installation**
  - [x] Add uninstall call after config loading and before module.Install()
  - [x] Ensure uninstall only runs when not in dry-run mode
  - [x] Add appropriate logging for cleanup phase
  - [x] Handle uninstall errors gracefully (log warning, continue)

- [x] **Add logging for two-phase installation process**
  - [x] Log start of cleanup phase
  - [x] Log completion of cleanup phase
  - [x] Log start of installation phase
  - [x] Ensure clear separation between phases in output

## Phase 2: Testing

- [x] **Add unit tests for install-uninstall integration**
  - [x] Test install calls uninstall when not in dry-run mode
  - [x] Test install skips uninstall in dry-run mode
  - [x] Test error handling when uninstall fails
  - [x] Test install with missing state file (uninstall should handle gracefully)

- [x] **Add integration tests for full cycle**
  - [x] Test install with existing previous installation
  - [x] Test install with no previous installation
  - [x] Test install with conflicting files after uninstall cleanup
  - [x] Test install with permission issues during uninstall

## Phase 3: Edge Cases and Error Handling

- [x] **Handle edge cases in uninstall-before-install flow**
  - [x] Test with corrupted state file during uninstall
  - [x] Test with permission denied during uninstall
  - [x] Test with partially successful uninstall
  - [x] Ensure installation continues even if uninstall partially fails

- [x] **Validate flag interactions**
  - [x] Test --force flag with uninstall-before-install
  - [x] Test --mkdir flag with uninstall-before-install
  - [x] Test --dry-run flag (should not run uninstall)
  - [x] Test flag combinations work correctly

## Phase 4: Documentation and Validation

- [x] **Update help text and documentation**
  - [x] Update install command help text to mention cleanup behavior
  - [x] Update any relevant README or documentation
  - [x] Ensure examples show new behavior

- [x] **Final validation and testing**
  - [x] Run full test suite to ensure no regressions
  - [x] Test manual installation scenarios
  - [x] Validate performance impact is minimal
  - [x] Test with various dotfile configurations

## Dependencies

- Core implementation depends on existing uninstall functionality
- Testing depends on existing test framework and patterns
- No new external dependencies required

## Notes

- Uninstall logic already exists and is well-tested
- Integration point is clean and minimal
- Error handling strategy is straightforward
- Backward compatibility is maintained automatically