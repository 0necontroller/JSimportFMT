# JSimportFMT

A high-performance CLI tool written in Go to automatically format and sort JavaScript and TypeScript imports by length. Optimized for very large repositories and CI/CD pipelines.

### Command

```bash
➜  express-api git:(main) jsimportfmt ./router/apps.ts  --dry-run
```
### Output
The output is all my imports are now formatted  from shortest to longest.    
Yes, Its a very niche tool and I just like my imports looking neet

```diff
--- ./router/apps.ts (original)
+++ ./router/apps.ts (formatted)
@@ -1,13 +1,13 @@
 import { Router } from "express";
-import { getTaxonomy } from "#/controllers/apps/taxonomy.Controller";
-import * as RfsController from "#/controllers/apps/rfs.Controller";
-import * as AppsController from "#/controllers/apps/apps.Controller";
-import * as ContributeController from "#/controllers/apps/contribute.Controller";
-import * as SubmitController from "#/controllers/apps/submit.Controller";
-import { authenticate } from "#/middleware/auth.middleware";
 import { UserRole } from "#/models/user.model";
+import { authenticate } from "#/middleware/auth.middleware";
 import { validate } from "#/middleware/validation.middleware";
 import * as ValidationSchema from "#/validation/apps.validation";
+import * as RfsController from "#/controllers/apps/rfs.Controller";
+import * as AppsController from "#/controllers/apps/apps.Controller";
+import { getTaxonomy } from "#/controllers/apps/taxonomy.Controller";
+import * as SubmitController from "#/controllers/apps/submit.Controller";
+import * as ContributeController from "#/controllers/apps/contribute.Controller";
 
 const router = Router();
 

✔ scanned 1 files
```

## Features

- **Blazing Fast**: Written in Go using concurrent worker pools.
- **Smart Parsing**: Uses a lightweight state-machine parser instead of a slow AST, ensuring comments, spacing, and formatting within imports are perfectly preserved.
- **Git & Gitignore Aware**: Automatically respects `.gitignore` rules if a `.git` folder is found within 2 levels of the target.
- **Non-Git Support**: Works seamlessly in any directory, even without a Git repository.
- **Type Separation**: Built-in support to separate type imports (`--separate-types`).
- **Interactive Mode**: Launch a guided session by simply running `jsimportfmt`.
- **Multiple Modes**: Supports Write (`--write`), Check (`--check`), and Dry-Run (`--dry-run`) modes.

## Installation

```bash
go install github.com/0necontroller/jsimportfmt@latest
```

## Build Instructions

```bash
git clone https://github.com/0necontroller/jsimportfmt.git
cd jsimportfmt
go build -o jsimportfmt .
```

## Usage Examples

### Interactive Mode
Launch a guided session by running the command without arguments:
```bash
jsimportfmt
```

### Format a Directory
Scan and rewrite all supported files in a folder:
```bash
jsimportfmt ./src --write
```

### Format a Single File
Target a specific file directly:
```bash
jsimportfmt ./src/App.tsx --write
```

### Dry-Run (Preview Changes)
See a unified diff of what would change without modifying any files:
```bash
jsimportfmt src --dry-run
```

**Example Output:**
```diff
--- src/index.ts (original)
+++ src/index.ts (formatted)
@@ -1,6 +1,6 @@
+import fs from "fs";
+import path from "path";
 import { App } from "./App";
 import { Config } from "./config";
-import path from "path";
-import fs from "fs";
```

### Separate Type Imports
Sort regular imports and type imports into separate blocks:
```bash
jsimportfmt src --write --separate-types
```

**Example:**
*Before:*
```ts
import type { User } from "./types";
import React from "react";
import type { Config } from "./config";
import { Button } from "./components";
```

*After:*
```ts
import React from "react";
import { Button } from "./components";

import type { User } from "./types";
import type { Config } from "./config";
```

### Check Ignore Rules
List the default ignores and those parsed from `.gitignore`:
```bash
jsimportfmt . --check-ignore
```

### Whitelist Directories
Allow formatting in ignored directories using a `.jifallow` file or the `--allow` flag.

**Automatic Persistence:**
Using the `--allow` flag (e.g., `--allow dist`) will automatically append the directory to your `.jifallow` file if it's not already listed. If the file doesn't exist, it will be created with a helpful header comment.

```bash
jsimportfmt . --allow dist --allow build
```

The `.jifallow` file supports `#` comments and one directory per line. Any directory listed there (or passed via `--allow`) will bypass both default ignores and `.gitignore` rules.

See [.jifallow.example](.jifallow.example) for a concrete example.

### Check Mode (CI)
Verify formatting without making changes:
```bash
jsimportfmt src --check
```

## CI Usage

You can use `jsimportfmt` in your CI pipeline to enforce import formatting. Use the `--check` flag:

```yaml
steps:
  - name: Check Import Formatting
    run: jsimportfmt . --check
```

### Exit Codes
- `0`: Success (no formatting needed, or all files successfully formatted in write mode)
- `1`: Formatting needed (returned in `--check` mode)
- `2`: Fatal error (e.g., file parsing failed or system error)

## Performance Notes
`jsimportfmt` streams file discovery and feeds a worker pool matched to your system's CPU count (`runtime.NumCPU()`). It reads, parses, and formats files incrementally, ensuring low memory overhead even for monorepos with tens of thousands of files.
