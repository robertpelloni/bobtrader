"""Patch config.go to add SecretsFile field and secrets loading logic."""

import os

script_dir = os.path.dirname(os.path.abspath(__file__))
config_path = os.path.join(script_dir, "..", "internal", "core", "config", "config.go")
config_path = os.path.normpath(config_path)

with open(config_path, "r", encoding="utf-8") as f:
    content = f.read()

# 1. Add SecretsFile field to AccountConfig
old_struct = '\tTestnet       bool     `json:"testnet"`\n}'
new_struct = (
    '\tTestnet       bool     `json:"testnet"`\n'
    '\tSecretsFile   string   `json:"secrets_file"` // path to JSON file with api_key/secret_key\n'
    "}\n"
    "\n"
    "type secretsFile struct {\n"
    '\tAPIKey    string `json:"api_key"`\n'
    '\tSecretKey string `json:"secret_key"`\n'
    "}"
)

if old_struct in content:
    content = content.replace(old_struct, new_struct, 1)
    print("Added SecretsFile field + secretsFile struct")
else:
    print("ERROR: old_struct not found")
    raise SystemExit(1)

# 2. Add secrets loading before the final return in Load()
old_return = "\treturn cfg, nil\n}"

new_return = (
    "\t// Load secrets from separate files if specified\n"
    "\tfor i := range cfg.Accounts {\n"
    "\t\tacct := &cfg.Accounts[i]\n"
    '\t\tif acct.SecretsFile != "" {\n'
    "\t\t\tsecData, err := os.ReadFile(acct.SecretsFile)\n"
    "\t\t\tif err != nil {\n"
    '\t\t\t\treturn Config{}, fmt.Errorf("read secrets file %s: %w", acct.SecretsFile, err)\n'
    "\t\t\t}\n"
    "\t\t\tvar sec secretsFile\n"
    "\t\t\tif err := json.Unmarshal(secData, &sec); err != nil {\n"
    '\t\t\t\treturn Config{}, fmt.Errorf("unmarshal secrets file %s: %w", acct.SecretsFile, err)\n'
    "\t\t\t}\n"
    '\t\t\tif sec.APIKey != "" {\n'
    "\t\t\t\tacct.APIKey = sec.APIKey\n"
    "\t\t\t}\n"
    '\t\t\tif sec.SecretKey != "" {\n'
    "\t\t\t\tacct.SecretKey = sec.SecretKey\n"
    "\t\t\t}\n"
    "\t\t}\n"
    "\t}\n"
    "\n"
    "\treturn cfg, nil\n"
    "}"
)

if old_return in content:
    content = content.replace(old_return, new_return, 1)
    print("Added secrets loading logic")
else:
    print("ERROR: old_return not found")
    raise SystemExit(1)

with open(config_path, "w", encoding="utf-8") as f:
    f.write(content)
print("Done")
