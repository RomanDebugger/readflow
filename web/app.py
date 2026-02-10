import streamlit as st
import os
import json
import subprocess
import base64

# Configuration
st.set_page_config(page_title="Readflow Intelligence", layout="wide")

# Directory setup
INPUT_DIR = "data/input_pdfs"
CHUNK_DIR = "data/chunks"

def run_backend():
    """Triggers the Go Pipeline"""
    with st.spinner("Readflow Go-Engine is extracting and scoring..."):
        result = subprocess.run(["go", "run", "src/main.go"], capture_output=True, text=True)
        if result.returncode == 0:
            st.success("Pipeline Complete!")
        else:
            st.error(f"Backend Error: {result.stderr}")

def display_pdf(file_path):
    """Shows the PDF in the UI"""
    with open(file_path, "rb") as f:
        base64_pdf = base64.b64encode(f.read()).decode('utf-8')
    pdf_display = f'<embed src="data:application/pdf;base64,{base64_pdf}" width="100%" height="800" type="application/pdf">'
    st.markdown(pdf_display, unsafe_allow_html=True)

# --- UI LAYOUT ---
st.title("🔍 Readflow | Document Intelligence")

# 1. File Uploader
uploaded_file = st.file_uploader("Drop a technical PDF here", type="pdf")

if uploaded_file:
    # Save the file to your Go input directory
    file_path = os.path.join(INPUT_DIR, uploaded_file.name)
    with open(file_path, "wb") as f:
        f.write(uploaded_file.getbuffer())
    
    # Run the Go pipeline automatically
    if st.button("Analyze Document Structure"):
        run_backend()

    # 2. Main Workspace (Split View)
    col1, col2 = st.columns([1, 1])

    with col1:
        st.subheader("Original Document")
        display_pdf(file_path)

    with col2:
        st.subheader("Intelligence Stream")
        
        # Load the resulting JSON
        json_path = os.path.join(CHUNK_DIR, uploaded_file.name.replace(".pdf", ".json"))
        
        if os.path.exists(json_path):
            with open(json_path, "r") as f:
                chunks = json.load(f)
            
            # Quality Filter Slider
            q_threshold = st.slider("Quality Threshold (Heuristic Filter)", 0.0, 1.0, 0.4)
            
            for c in chunks:
                if c['quality'] >= q_threshold:
                    # Style based on Type
                    bg_color = "#f0f2f6" if c['type'] == 'paragraph' else "#e1f5fe"
                    border_color = "#1E88E5" if c['type'] == 'title' else "#cfd8dc"
                    
                    st.markdown(f"""
                        <div style="background-color:{bg_color}; border-left: 5px solid {border_color}; padding:15px; margin-bottom:10px; border-radius:5px; color: black;">
                            <small style="color:gray;">Page {c['page']} | Score: {c['quality']:.2f}</small>
                            <div style="font-weight:{'bold' if c['type'] == 'title' else 'normal'}; font-size:{'20px' if c['type'] == 'title' else '16px'};">
                                {c['text']}
                            </div>
                        </div>
                    """, unsafe_allow_html=True)
        else:
            st.info("Run the Analysis to see extracted intelligence.")