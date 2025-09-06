# Test Data Reference

## Overview

The usacloud-update project uses golden file testing to ensure transformation accuracy. This document describes the test data structure and provides examples of the transformation process.

## Test Data Structure

### Input Test File: `testdata/sample_v0_v1_mixed.sh`

This file contains representative examples of usacloud commands from different versions (v0.x, v1.0) that need transformation to v1.1 compatibility.

**File Structure**:
```bash
#!/usr/bin/env bash
set -euo pipefail

# Various usacloud command examples representing common migration scenarios
```

### Expected Output: `testdata/expected_v1_1.sh`

Contains the expected transformation results, including:
- Transformed commands
- Explanatory comments with reasons and documentation links
- Generated header indicating automatic transformation

## Transformation Examples

### 1. Output Format Migration

**Input**:
```bash
# v0風: csv/tsv
usacloud server list --output-type=csv
```

**Output**:
```bash
# v0風: csv/tsv  
usacloud server list --output-type=json # usacloud-update: v1.0でcsv/tsvは廃止。jsonに置換し、必要なら --query/jq を利用してください (https://docs.usacloud.jp/usacloud/upgrade/v1_0_0/)
```

**Transformation Details**:
- **Pattern**: `--output-type=(csv|tsv)`
- **Replacement**: `--output-type=json`
- **Reason**: CSV/TSV output deprecated in v1.0
- **Alternative**: Use JSON with --query/jq for filtering

### 2. Selector Deprecation

**Input**:
```bash
# v0風: selector
usacloud disk read --selector name=mydisk
usacloud server delete --selector tag=to-be-removed
```

**Output**:
```bash
# v0風: selector
usacloud disk read mydisk # usacloud-update: --selectorはv1で廃止。ID/名称/タグをコマンド引数に指定する仕様へ移行 (https://docs.usacloud.jp/usacloud/upgrade/v1_0_0/)
usacloud server delete to-be-removed # usacloud-update: --selectorはv1で廃止。ID/名称/タグをコマンド引数に指定する仕様へ移行 (https://docs.usacloud.jp/usacloud/upgrade/v1_0_0/)
```

**Transformation Details**:
- **Pattern**: `--selector value`
- **Replacement**: `value` (as command argument)
- **Reason**: --selector flag removed in v1
- **Migration**: Use direct command arguments

### 3. Resource Name Changes

**Input**:
```bash
# v0風: リソース名
usacloud iso-image list
usacloud startup-script list
usacloud ipv4 read --zone tk1a --ipaddress 203.0.113.10
```

**Output**:
```bash
# v0風: リソース名
usacloud cdrom list # usacloud-update: v1ではリソース名がcdromに統一 (https://manual.sakura.ad.jp/cloud-api/1.1/cdrom/index.html)
usacloud note list # usacloud-update: v1ではstartup-scriptはnoteに統一 (https://docs.usacloud.jp/usacloud/)
usacloud ipaddress read --zone tk1a --ipaddress 203.0.113.10 # usacloud-update: v1ではIPv4関連はipaddressに整理 (https://docs.usacloud.jp/usacloud/references/ipaddress/)
```

**Transformation Details**:
- `iso-image` → `cdrom`: Resource name standardization
- `startup-script` → `note`: Unified naming convention
- `ipv4` → `ipaddress`: IPv4 functionality consolidated

### 4. Product Alias Cleanup

**Input**:
```bash
# v0: product-*
usacloud product-disk list
```

**Output**:
```bash
# v0: product-*
usacloud disk-plan list # usacloud-update: v1系では *-plan へ名称統一 (https://docs.usacloud.jp/usacloud/)
```

**Transformation Details**:
- **Pattern**: `product-*` aliases
- **Replacement**: `*-plan` naming convention
- **Reason**: Consistent naming in v1 series

### 5. Command Deprecation

**Input**:
```bash
# 廃止コマンド
usacloud summary
```

**Output**:
```bash
# 廃止コマンド
# usacloud summary # usacloud-update: summaryコマンドはv1で廃止。要件に応じて bill/self/各list か rest を利用してください (https://docs.usacloud.jp/usacloud/upgrade/v1_0_0/)
```

**Transformation Details**:
- **Action**: Comment out deprecated commands
- **Reason**: Command removed in v1
- **Alternatives**: Use bill/self/list commands or REST API

### 6. Service Deprecation

**Input**:
```bash
# 非サポート(object-storage)
usacloud object-storage list
```

**Output**:
```bash
# 非サポート(object-storage)
# usacloud object-storage list # usacloud-update: v1ではオブジェクトストレージ操作は非対応方針。S3互換ツール/他プロバイダやTerraformを検討 (https://github.com/sacloud/usacloud/issues/585)
```

**Transformation Details**:
- **Action**: Comment out object-storage commands
- **Reason**: Not supported in v1
- **Alternatives**: Use S3-compatible tools or Terraform

### 7. Parameter Normalization

**Input**:
```bash
# v1.0以降: allゾーン
usacloud server list --zone = all
```

**Output**:
```bash
# v1.0以降: allゾーン
usacloud server list --zone=all # usacloud-update: 全ゾーン一括操作は --zone=all を推奨 (https://docs.usacloud.jp/usacloud/upgrade/v1_0_0/)
```

**Transformation Details**:
- **Pattern**: `--zone = all` (with spaces)
- **Replacement**: `--zone=all` (without spaces)
- **Reason**: Standardized parameter format

## Generated Header

Every transformed file begins with:
```bash
# Updated for usacloud v1.1 by usacloud-update — DO NOT EDIT ABOVE THIS LINE
```

**Purpose**:
- Identifies automatically transformed files
- Warns against manual edits above the line
- Provides tool attribution

## Testing Process

### Golden File Test Flow

1. **Input Processing**: Read `sample_v0_v1_mixed.sh`
2. **Transformation**: Apply all rules via transformation engine
3. **Comparison**: Compare output with `expected_v1_1.sh`
4. **Validation**: Byte-for-byte matching ensures accuracy

### Test Execution

```bash
# Run tests
go test ./...

# Update golden files (after rule changes)
go test -run Golden -update ./...

# Manual verification
make verify-sample
```

### Test Coverage

The test data covers all transformation categories:
- ✅ Output format migration (csv/tsv → json)
- ✅ Selector deprecation (--selector → arguments)
- ✅ Resource renaming (iso-image, startup-script, ipv4)
- ✅ Product alias cleanup (product-* → *-plan)
- ✅ Command deprecation (summary)
- ✅ Service deprecation (object-storage)
- ✅ Parameter normalization (--zone spacing)

## Adding New Test Cases

### Process for New Transformations

1. **Add Input Example**: Update `sample_v0_v1_mixed.sh`
2. **Add Rule**: Implement transformation in `ruledefs.go`
3. **Update Golden**: Run `go test -run Golden -update`
4. **Verify**: Check that `expected_v1_1.sh` contains expected output
5. **Commit**: Include both test data and rule changes

### Test Data Guidelines

- **Representative Examples**: Use realistic usacloud command patterns
- **Edge Cases**: Include boundary conditions and special cases
- **Documentation**: Comment test cases with their purpose
- **Categorization**: Group similar transformation types together

## Validation and Debugging

### Manual Verification
```bash
# Process sample file manually
make run

# Compare with expected output  
diff testdata/expected_v1_1.sh /tmp/out.sh
```

### Debugging Transformation Issues

1. **Check Input**: Verify test case matches intended pattern
2. **Rule Testing**: Test regex patterns in isolation
3. **Golden Update**: Regenerate expected output after fixes
4. **Visual Diff**: Use diff tools to compare expected vs actual output

## Maintenance

### Keeping Test Data Current

- **Regular Updates**: Update test cases when adding new rules
- **Real-world Examples**: Base test cases on actual user scripts
- **Documentation Links**: Ensure URLs in comments remain valid
- **Version Compatibility**: Update version references as needed

### Golden File Management

- **Source Control**: Commit golden files alongside code changes
- **Review Process**: Review golden file changes during code review
- **Backward Compatibility**: Ensure changes don't break existing transformations