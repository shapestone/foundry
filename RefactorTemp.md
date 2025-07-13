# Foundry Architecture Refactoring - Remaining Phases

## **Phase 2: Parser & Input Validation** (Week 2)

### 🎯 **Objective**: Extract CLI parsing and validation logic

### **Tasks**:
1. **Create `internal/parser/` package**
    - `interfaces.go` - Parser contracts
    - `handler_parser.go` - Convert CLI args to `HandlerSpec`
    - `model_parser.go` - Convert CLI args to `ModelSpec`
    - `validation.go` - Input validation logic

2. **Extract validation from commands**
    - Move `ValidateComponentName` to parser
    - Create structured validation with rich errors
    - Add context-aware validation rules

3. **Create comprehensive parser tests**
    - `test/parser/` directory
    - Unit tests for all parsers
    - Validation edge cases
    - Error message quality tests

4. **Refactor 2-3 commands to use parser**
    - Update handler command to use parser
    - Update model command to use parser
    - Demonstrate the pattern

### **Expected Coverage**: 55% → 65%

---

## **Phase 3: Template System** (Week 3)

### 🎯 **Objective**: Formalize and externalize template system

### **Tasks**:
1. **Create `internal/templates/` package**
    - `interfaces.go` - Template contracts
    - `loader.go` - Template loading from filesystem
    - `renderer.go` - Template rendering engine
    - `cache.go` - Template caching for performance

2. **Move templates to external files**
    - `templates/` directory structure
    - `templates/handler/` - Handler templates
    - `templates/model/` - Model templates
    - `templates/middleware/` - Middleware templates

3. **Template testing**
    - `test/templates/` directory
    - Template rendering tests
    - Template validation tests
    - Performance benchmarks

4. **Integration with scaffolder**
    - Update adapters to use new template system
    - Remove template hardcoding
    - Add template discovery

### **Expected Coverage**: 65% → 75%

---

## **Phase 4: Command Refactoring** (Week 4)

### 🎯 **Objective**: Refactor all remaining commands

### **Tasks**:
1. **Refactor all commands to use new architecture**
    - `cmd/foundry/cmd/handler.go` → use scaffolder
    - `cmd/foundry/cmd/model.go` → use scaffolder
    - `cmd/foundry/cmd/middleware.go` → use scaffolder
    - `cmd/foundry/cmd/database.go` → use scaffolder
    - `cmd/foundry/cmd/wire.go` → use scaffolder

2. **Implement remaining scaffolder functions**
    - Complete `CreateModel` implementation
    - Complete `CreateMiddleware` implementation
    - Complete `CreateDatabase` implementation
    - Complete `WireHandler` implementation

3. **Command integration tests**
    - End-to-end command testing
    - Real CLI interaction tests
    - Error handling verification

4. **Clean up old code**
    - Remove duplicate logic
    - Consolidate utilities
    - Update documentation

### **Expected Coverage**: 75% → 85%

---

## **Phase 5: Advanced Features** (Week 5+)

### 🎯 **Objective**: Polish and advanced capabilities

### **Tasks**:
1. **Enhanced CLI Experience**
    - Better error messages with suggestions
    - Interactive mode for complex operations
    - Progress indicators for long operations
    - Colored output and better formatting

2. **Advanced Scaffolding Features**
    - Template customization
    - Project-specific configuration
    - Custom field types for models
    - Relationship scaffolding

3. **Performance & Reliability**
    - Caching for better performance
    - Atomic operations with rollback
    - Concurrent operation support
    - Memory optimization

4. **Developer Experience**
    - Plugin system foundation
    - Configuration file support
    - Auto-completion for shells
    - IDE integration helpers

### **Expected Coverage**: 85%+ → 90%+

---

## **Future Enhancements** (v0.2+)

### **Plugin System**
- External plugin loading
- Plugin discovery and management
- Community template marketplace
- Custom scaffolding rules

### **Advanced Templates**
- Conditional template logic
- Template inheritance
- Multi-file template packages
- Template version management

### **Enterprise Features**
- Team configuration sharing
- Enterprise template libraries
- Audit logging
- Integration with CI/CD

---

## **Coverage Progression Summary**

| Phase | Focus | Expected Coverage | Key Improvement |
|-------|-------|------------------|----------------|
| **Current** | Foundation | 46.2% | Scaffolder architecture |
| **Phase 2** | Parser | 65% | CLI input validation |
| **Phase 3** | Templates | 75% | External template system |
| **Phase 4** | Commands | 85% | All commands refactored |
| **Phase 5** | Polish | 90%+ | Advanced features |

## **Success Metrics by Phase**

### **Phase 2 Success**:
- [ ] Parser package with 90%+ coverage
- [ ] 3 commands using parser
- [ ] Rich validation error messages
- [ ] Overall coverage: 65%

### **Phase 3 Success**:
- [ ] External template files
- [ ] Template caching working
- [ ] Template testing framework
- [ ] Overall coverage: 75%

### **Phase 4 Success**:
- [ ] All commands refactored
- [ ] No more 0% coverage functions
- [ ] Integration tests passing
- [ ] Overall coverage: 85%

### **Phase 5 Success**:
- [ ] Advanced CLI features
- [ ] Plugin system foundation
- [ ] Performance optimizations
- [ ] Overall coverage: 90%+

---

**Each phase builds on the previous one, maintaining the clean architecture principles established in Phase 1. The incremental approach ensures you can validate each step before proceeding to the next.**

Ready to tackle Phase 2? 🚀