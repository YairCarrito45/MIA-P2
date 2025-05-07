import React, { useState, useEffect, useRef } from "react";
import Editor from "@monaco-editor/react";
import "./App.css";
import LoginForm from "./LoginForm";
import { Routes, Route, useNavigate } from "react-router-dom";
import DiskSelector from "./DiskSelector";
import PartitionSelector from "./PartitionSelector";
import FileViewer from "./FileViewer";

function App() {
  const [commandInput, setCommandInput] = useState("");
  const [output, setOutput] = useState("");
  const textareaRef = useRef(null);
  const [usuarioActual, setUsuarioActual] = useState(null);
  const [mostrarLogin, setMostrarLogin] = useState(false);
  const navigate = useNavigate();

  useEffect(() => {
    const savedUser = localStorage.getItem("user");
    if (savedUser) {
      try {
        setUsuarioActual(JSON.parse(savedUser));
      } catch {
        localStorage.removeItem("user");
      }
    }
  }, []);

  useEffect(() => {
    const textarea = textareaRef.current;
    if (textarea) {
      textarea.style.height = "auto";
      textarea.style.height = `${textarea.scrollHeight}px`;
    }
  }, [output]);

  const handleFileUpload = (e) => {
    const file = e.target.files[0];
    if (file && file.name.endsWith(".smia")) {
      const reader = new FileReader();
      reader.onload = (event) => {
        setCommandInput(event.target.result);
      };
      reader.readAsText(file);
    } else {
      alert("Por favor selecciona un archivo .smia válido");
    }
  };

  const handleExecute = async () => {
    try {
      const response = await fetch("http://localhost:8080/api/analizar", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          text: commandInput,
        }),
      });
      const result = await response.json();
      setOutput(result.output || "Sin salida.");


    } catch {
      setOutput("Error al comunicarse con el backend.");
    }
  };

  const handleClear = () => {
    setCommandInput("");
    setOutput("");
  };

  const handleLogout = () => {
    localStorage.removeItem("user");
    setUsuarioActual(null);
    setCommandInput("");
    setOutput("");
  };

  return (
    <Routes>
      <Route
        path="/"
        element={
          <div className="App">
            <header className="App-header">
              <h1>Sistema de Archivos EXT2/EXT3 - Proyecto MIA</h1>
              {usuarioActual && (
                <p>
                  Sesión activa: <strong>{usuarioActual.username}</strong> (ID: {usuarioActual.partitionId})
                </p>
              )}
            </header>

            <div className="controls">
              <input type="file" accept=".smia" onChange={handleFileUpload} />
              <button onClick={handleExecute}>Ejecutar</button>
              <button onClick={handleClear}>Limpiar</button>
              {!usuarioActual ? (
                <button onClick={() => setMostrarLogin(true)} style={{ backgroundColor: "#2e7d32", color: "white" }}>
                  Iniciar Sesión
                </button>
              ) : (
                <button onClick={handleLogout} style={{ backgroundColor: "#880e4f", color: "white" }}>
                  Cerrar Sesión ({usuarioActual.username})
                </button>
              )}
            </div>

            <div className="editor-container">
              <div className="editor">
                <label>Entrada:</label>
                <Editor
                  height="560px"
                  language="plaintext"
                  theme="hc-black"
                  value={commandInput}
                  onChange={(value) => setCommandInput(value)}
                  options={{
                    minimap: { enabled: false },
                    fontSize: 14,
                    lineNumbers: "on",
                    scrollBeyondLastLine: false,
                    wordWrap: "on",
                  }}
                />
              </div>

              <div className="editor">
                <label>Salida:</label>
                <textarea
                  ref={textareaRef}
                  className="salida"
                  value={output}
                  readOnly
                  placeholder="#------Estiben Yair Lopez Leveron------ 202204578----"
                />
              </div>
            </div>

            {mostrarLogin && (
              <>
                <div
                  style={{
                    position: "fixed",
                    top: 0,
                    left: 0,
                    width: "100vw",
                    height: "100vh",
                    backgroundColor: "rgba(0, 0, 0, 0.5)",
                    zIndex: 999,
                  }}
                />
                <div
                  style={{
                    position: "fixed",
                    top: "30%",
                    left: "50%",
                    transform: "translate(-50%, -50%)",
                    background: "#ffffff",
                    padding: "2rem",
                    borderRadius: "10px",
                    boxShadow: "0 0 15px rgba(0,0,0,0.3)",
                    zIndex: 1000,
                  }}
                >
                  <LoginForm
                    onLogin={(info) => {
                      setUsuarioActual(info);
                      localStorage.setItem("user", JSON.stringify(info));
                      setMostrarLogin(false);
                      navigate("/discos");
                    }}
                  />

                  <button
                    onClick={() => setMostrarLogin(false)}
                    style={{
                      marginTop: "1rem",
                      backgroundColor: "#999",
                      color: "#fff",
                      padding: "0.4rem 1rem",
                      border: "none",
                      borderRadius: "5px",
                      cursor: "pointer",
                    }}
                  >
                    Cancelar
                  </button>
                </div>
              </>
            )}
          </div>
        }
      />
      <Route path="/discos" element={<DiskSelector />} />
      <Route path="/partitions/:diskName" element={<PartitionSelector />} />
      <Route path="/viewer/:id" element={<FileViewer />} />
    </Routes>
  );
}

export default App;
