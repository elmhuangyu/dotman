## 1. Configuration File Support
- [ ] 1.1 Create config file parser for YAML format
- [ ] 1.2 Define Config struct with TargetDir field
- [ ] 1.3 Implement config file discovery in module directories

## 2. Module Discovery
- [ ] 2.1 Implement module directory scanning logic
- [ ] 2.2 Add logic to detect "Dotfile" config in each module
- [ ] 2.3 Integrate with existing dotfiles directory structure

## 3. Install Command Enhancement
- [ ] 3.1 Update install command to read module configs
- [ ] 3.2 Implement file mapping based on target_dir from config
- [ ] 3.3 Add logging for config file discovery and usage

## 4. Uninstall Command Enhancement
- [ ] 4.1 Update uninstall command to use module configs
- [ ] 4.2 Track files installed using config-based targets
- [ ] 4.3 Ensure proper cleanup of config-managed installations

## 5. Verify Command Enhancement
- [ ] 5.1 Update verify command to check config-based installations
- [ ] 5.2 Validate that files are correctly linked to target locations
- [ ] 5.3 Add detailed reporting for config-based verification

## 6. Error Handling and Validation
- [ ] 6.1 Add validation for config file format
- [ ] 6.2 Handle missing or invalid config files gracefully
- [ ] 6.3 Provide clear error messages for config-related issues

## 7. Documentation and Tests
- [ ] 7.1 Update README with config file usage examples
- [ ] 7.2 Add unit tests for config parsing logic
- [ ] 7.3 Add integration tests for module-based installation