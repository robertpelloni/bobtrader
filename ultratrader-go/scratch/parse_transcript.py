import json

def parse_transcript():
    file_path = r"C:\Users\hyper\.gemini\antigravity\brain\a0b536be-ba1a-4a16-b4cd-f81b82c56dd1\.system_generated\logs\transcript.jsonl"
    with open(file_path, "r", encoding="utf-8") as f:
        for line in f:
            if not line.strip():
                continue
            try:
                data = json.loads(line)
                source = data.get("source")
                step_type = data.get("type")
                step_idx = data.get("step_index")
                
                if step_type == "USER_INPUT":
                    content = data.get("content", "")
                    print(f"[{step_idx}] USER: {content.strip()}")
                elif step_type == "PLANNER_RESPONSE":
                    content = data.get("content", "")
                    # print first 100 chars of model response
                    snippet = content.strip().replace("\n", " ")[:150]
                    print(f"[{step_idx}] AGENT: {snippet}...")
            except Exception as e:
                pass

if __name__ == "__main__":
    parse_transcript()
