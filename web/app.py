import os
import json
import subprocess
import google.generativeai as genai
from fastapi import FastAPI, UploadFile, File, HTTPException
from fastapi.responses import FileResponse
import os
import json
import requests
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
model = genai.GenerativeModel('gemini-2.5-flash-lite')

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

# Grab the internal URL you set in Railway variables
ENGINE_URL = os.getenv("ENGINE_URL", "http://readflow.railway.internal:8080")

@app.post("/api/process")
async def process_pdf(file: UploadFile = File(...)):
    filename = get_safe_filename(file)
    
    # 1. Read the file into memory (DO NOT save it to disk)
    file_bytes = await file.read()
    
    # 2. Forward the file directly to the Go Engine via HTTP
    try:
        go_response = requests.post(
            f"{ENGINE_URL}/process", 
            files={"file": (filename, file_bytes, file.content_type or "application/pdf")}
        )
        
        # 3. Return whatever JSON the Go engine gives us back to React
        if go_response.status_code == 200:
            return go_response.json()
        else:
            print(f"Go Engine Error: {go_response.text}")
            raise HTTPException(status_code=500, detail="Go Engine failed to process the PDF")
            
    except requests.exceptions.ConnectionError:
        raise HTTPException(status_code=500, detail="Could not connect to the Go Engine")

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

@app.get("/{full_path:path}")
async def serve_react(full_path: str):
    # Check if the browser is asking for a specific file in dist (like favicon.svg)
    static_file = os.path.join("dist", full_path)
    if os.path.exists(static_file) and os.path.isfile(static_file):
        return FileResponse(static_file)
    
    # If not a specific file, serve the main React index.html
    return FileResponse("dist/index.html")