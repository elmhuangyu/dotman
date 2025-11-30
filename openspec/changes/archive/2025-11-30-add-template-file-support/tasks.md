## 1. Template File Detection
- [x] 1.1 Add function to detect `.dot-tmpl` files in file mapping
- [x] 1.2 Update file mapping to track template vs regular files

## 2. Template Processing
- [x] 2.1 Add Go text template rendering function using root_config.Vars
- [x] 2.2 Integrate template processing into install workflow
- [x] 2.3 Handle template rendering errors appropriately

## 3. File Generation
- [x] 3.1 Add function to generate rendered files from templates
- [x] 3.2 Update install logic to generate files instead of symlinks for templates
- [x] 3.3 Ensure proper file permissions and directory creation

## 4. Testing
- [x] 4.1 Create test cases for template file detection
- [x] 4.2 Create test cases for template rendering with various variables
- [x] 4.3 Create test cases for file generation vs symlink creation
- [x] 4.4 Add integration tests for full template workflow

## 5. Validation
- [x] 5.1 Update validation logic to handle template files
- [x] 5.2 Ensure template syntax validation during dry-run
- [x] 5.3 Handle missing template variables gracefully