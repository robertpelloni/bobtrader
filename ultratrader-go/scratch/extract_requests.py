import json

def extract():
    file_path = r"C:\Users\hyper\.gemini\antigravity\brain\a0b536be-ba1a-4a16-b4cd-f81b82c56dd1\.system_generated\logs\transcript.jsonl"
    out_path = r"C:\Users\hyper\.gemini\antigravity\brain\a0b536be-ba1a-4a16-b4cd-f81b82c56dd1\user_requests.md"
    
    requests = []
    with open(file_path, "r", encoding="utf-8") as f:
        for line in f:
            if not line.strip():
                continue
            try:
                data = json.loads(line)
                if data.get("type") == "USER_INPUT":
                    content = data.get("content", "").strip()
                    idx = data.get("step_index")
                    created_at = data.get("created_at", "")
                    requests.append(f"### Step {idx} ({created_at})\n\n{content}\n\n---\n")
            except Exception:
                pass
                
    with open(out_path, "w", encoding="utf-8") as f:
        f.write("# Conversation User Requests History\n\n")
        f.write("".join(requests))
    print(f"Successfully wrote {len(requests)} requests to {out_path}")

if __name__ == "__main__":
    extract()
