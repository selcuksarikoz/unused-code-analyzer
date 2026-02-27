# Test Samples

This directory contains test files for the Unused Code Analyzer extension.

## Directory Structure

```
test-samples/
├── javascript/         # Plain JavaScript test files (.js)
│   ├── test.js              # Basic JS test
│   ├── modern-js.js         # ES6+ features test
│   └── test-wasm*.js        # WASM-related tests
├── typescript/         # TypeScript test files (.ts)
│   ├── main.ts              # Basic TS test
│   ├── feature.ts           # Feature test
│   ├── unused-imports.ts    # Import patterns
│   ├── unused-variables.ts  # Variable patterns
│   ├── unused-parameters.ts # Parameter patterns
│   ├── type-imports.ts      # Type-only imports
│   ├── mixed-test.ts        # Mixed scenarios
│   └── advanced-types.ts    # Advanced TypeScript types
├── react/              # React/TSX test files (.jsx, .tsx)
│   ├── main.jsx             # JSX test
│   ├── react-component.tsx  # Functional component
│   ├── hooks-test.tsx       # React hooks
│   └── class-components.tsx # Class components
├── vue/                # Vue test files (.vue)
│   ├── component.vue        # Vue 3 + TypeScript
│   └── vue2-component.vue   # Vue 2 Options API
├── svelte/             # Svelte test files (.svelte)
│   └── component.svelte     # Svelte + TypeScript
├── astro/              # Astro test files (.astro)
│   └── page.astro           # Astro page
├── shared/             # Shared utility modules
│   ├── utils.ts             # Shared utilities
│   └── reexports.ts         # Re-export examples
└── other/              # Other languages (Go, Python, PHP, Ruby)
    ├── main.go
    ├── main.py
    ├── main.php
    ├── main.rb
    └── goproject/
```

## Test Categories

### JavaScript Tests
- Basic variable declarations (const, let, var)
- Function declarations and expressions
- Class definitions
- Import/Export patterns
- Modern ES6+ features

### TypeScript Tests
- Type-only imports (`import type`)
- Interfaces and type aliases
- Generics
- Enums
- Advanced types (mapped, conditional, unions)

### React Tests
- Functional components
- Class components
- React hooks (useState, useEffect, useCallback, etc.)
- JSX patterns
- Props and state

### Framework Tests (Vue, Svelte, Astro)
- Component structure
- Props and emits
- Lifecycle hooks
- Reactive state
- Template usage

## Expected Unused Detections

Each test file contains comments marking:
- `// Unused...` - Should be flagged as unused
- `// Used...` - Should NOT be flagged
- `// should NOT be flagged` - Intentionally used

## Running Tests

1. Open this workspace in VS Code
2. Start the extension (F5)
3. Run "Scan Workspace" command
4. Check the results panel for detected unused code
