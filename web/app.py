import os
import json
import subprocess
import google.generativeai as genai
from fastapi import FastAPI, UploadFile, File, HTTPException
from fastapi.responses import FileResponse
from fastapi.middleware.cors import CORSMiddleware
from dotenv import load_dotenv
from fastapi.staticfiles import StaticFiles
from pydantic import BaseModel
from typing import List


load_dotenv()
genai.configure(api_key=os.getenv("GEMINI_API_KEY"))
model = genai.GenerativeModel('gemini-2.5-flash')

app = FastAPI()

app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_methods=["*"],
    allow_headers=["*"],
)

class ChatRequest(BaseModel):
    query: str
    context: List[dict]
    
def get_safe_filename(file: UploadFile) -> str:
    name = file.filename
    if name is None:
        raise HTTPException(status_code=400, detail="No filename provided")
    return name

@app.post("/api/process")
async def process_pdf(file: UploadFile = File(...)):
    filename = get_safe_filename(file)
    
    # Pathing relative to /web directory
    input_path = os.path.join("..", "data", "input_pdfs", filename)
    
    with open(input_path, "wb") as f:
        f.write(await file.read())
    
    # 2. Trigger the Go Engine (already modularized in /src)
    try:
        subprocess.run(["go", "run", "../src/main.go"], check=True)
    except subprocess.CalledProcessError:
        raise HTTPException(status_code=500, detail="Go Engine Failed")
    
    # 3. Load result from /data/chunks/
    json_name = filename.rsplit('.', 1)[0] + ".json"
    json_path = os.path.join("..", "data", "chunks", json_name)
    
    if os.path.exists(json_path):
        with open(json_path, "r") as f:
            return json.load(f)
    
    return {"error": "Refined data not found"}

@app.post("/api/chat")
async def chat(request: ChatRequest):
    formatted_context = "\n".join([
        f"[Page {c.get('page', 'Unknown')}]: {c.get('text', '')}" 
        for c in request.context
    ])
    
    prompt = f"""You are a document expert. 
    Context: {formatted_context}
    User: {request.query}
    Cite specific page numbers in your answer."""
    
    try:
        response = model.generate_content(prompt)
        return {"response": response.text}
    except Exception as e:
        print(f"Gemini Error: {e}")
        raise HTTPException(status_code=500, detail="Gemini failed to process request")

if os.path.exists("dist/assets"):
    app.mount("/assets", StaticFiles(directory="dist/assets"), name="static_assets")

# 2. Catch-all route to serve the React UI
@app.get("/{full_path:path}")
async def serve_react(full_path: str):
    # Check if the browser is asking for a specific file in dist (like favicon.svg)
    static_file = os.path.join("dist", full_path)
    if os.path.exists(static_file) and os.path.isfile(static_file):
        return FileResponse(static_file)
    
    # If not a specific file, serve the main React index.html
    return FileResponse("dist/index.html")