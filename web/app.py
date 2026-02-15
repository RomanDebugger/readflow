import streamlit as st
import json
import os
import base64
import subprocess
from dotenv import load_dotenv
import requests
env_path = os.path.join(os.path.dirname(__file__), '..', '.env')
load_dotenv(dotenv_path=env_path)

st.set_page_config(page_title="Readflow v1.0", layout="wide")

st.markdown("""
    <style>
    .main { background-color: #0e1117; color: #e0e0e0; }
    .nav-card {
        background-color: #161b22;
        padding: 12px;
        border-radius: 8px;
        border-left: 5px solid #00d4ff;
        margin-bottom: 10px;
        transition: 0.3s;
    }
    .nav-card:hover { border-left: 5px solid #ffffff; background-color: #1c2128; }
    .stSlider [data-baseweb="slider"] { margin-top: 20px; }
    </style>
""", unsafe_allow_html=True)

INPUT_DIR = "data/input_pdfs"
CHUNK_DIR = "data/chunks"

def get_pdf_viewer(file_path, page):
    """Embeds PDF with page-specific anchoring"""
    with open(file_path, "rb") as f:
        base64_pdf = base64.b64encode(f.read()).decode('utf-8')
    return f'<embed src="data:application/pdf;base64,{base64_pdf}#page={page}" width="100%" height="950" type="application/pdf">'

with st.sidebar:
    st.title("Readflow Engine")
    uploaded_file = st.file_uploader("Drop Technical PDF", type="pdf")
    
    q_threshold = st.slider("Signal Sensitivity (Quality Filter)", 0.0, 1.0, 0.6)
    st.caption("Lower: Show more context | Higher: Show structural anchors only")

    if uploaded_file:
        save_path = os.path.join(INPUT_DIR, uploaded_file.name)
        with open(save_path, "wb") as f:
            f.write(uploaded_file.getbuffer())
        
        if st.button("Re-Analyze Structure", use_container_width=True):
            subprocess.run(["go", "run", "src/main.go"])
            st.rerun()

if uploaded_file:
    json_path = os.path.join(CHUNK_DIR, uploaded_file.name.replace(".pdf", ".json"))
    
    if os.path.exists(json_path):
        with open(json_path, "r") as f:
            data = json.load(f)

        col_pdf, col_intel = st.columns([1.5, 1])

        with col_pdf:
            target_page = st.session_state.get("page", 1)
            st.markdown(get_pdf_viewer(save_path, target_page), unsafe_allow_html=True)

        with col_intel:
            tab_nav, tab_chat = st.tabs(["📍 Semantic Navigator", "🤖 AI Expert"])

            with tab_nav:
                st.subheader("Gated Structural Anchors")
                filtered_chunks = [c for c in data if c['quality'] >= q_threshold]
                st.info(f"Showing {len(filtered_chunks)} high-signal units.")

                for c in filtered_chunks:
                    with st.container():
                        label = f"[{c['type'].upper()}] Page {c['page']}"
                        if st.button(f"{label}: {c['text'][:60]}...", key=c['chunk_id']):
                            st.session_state.page = c['page']
                            st.rerun()
                        st.markdown('</div>', unsafe_allow_html=True)

            with tab_chat:
                st.subheader("Intelligent Query")
                
                if "messages" not in st.session_state:
                    st.session_state.messages = []

                for message in st.session_state.messages:
                    with st.chat_message(message["role"]):
                        st.markdown(message["content"])

                user_query = st.chat_input("Query high-signal context...")
                
                if user_query:
                    st.session_state.messages.append({"role": "user", "content": user_query})
                    with st.chat_message("user"):
                        st.markdown(user_query)

                    context_text = "\n".join([f"PAGE {c['page']}: {c['text']}" for c in filtered_chunks])
                    
                    with st.chat_message("assistant"):
                        with st.spinner("Ollama is reconstructing and analyzing..."):
                            try:
                                response = requests.post(
                                    "http://localhost:11434/api/generate",
                                    json={
                                        "model": "gemma3:latest",
                                        "prompt": f"""System: You are a document expert. 
                                        The following context has extraction artifacts like missing spaces. 
                                        Reconstruct the meaning, answer the question accurately, and cite the PAGE number.
                                        
                                        CONTEXT:
                                        {context_text}
                                        
                                        USER QUESTION:
                                        {user_query}""",
                                        "stream": False
                                    }
                                )
                                full_response = response.json().get('response', "Error: No response from local model.")
                                st.markdown(full_response)
                                st.session_state.messages.append({"role": "assistant", "content": full_response})
                            except Exception as e:
                                st.error(f"Ollama Connection Failed. Is 'ollama serve' running? \nError: {e}")
    else:
        st.warning("Go-Engine Analysis Required.")