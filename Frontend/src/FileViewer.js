import React, { useEffect, useState } from "react";
import { useParams, useNavigate } from "react-router-dom";

function FileViewer() {
  const { id } = useParams();
  const [tree, setTree] = useState(null);
  const navigate = useNavigate();

  useEffect(() => {
    fetch(`http://localhost:3001/filesystem/${id}`)
      .then((res) => res.json())
      .then((data) => setTree(data))
      .catch(() => setTree(null));
  }, [id]);

  const renderNode = (node, depth = 0) => {
    const isFolder = node.type === "folder";
    return (
      <div key={node.name + node.path} style={{ marginLeft: depth * 20 }}>
        <p style={{ fontWeight: isFolder ? "bold" : "normal" }}>
          {isFolder ? "📁" : "📄"} {node.name}
        </p>
        {isFolder && node.children && node.children.map((child) => renderNode(child, depth + 1))}
      </div>
    );
  };

  return (
    <div style={styles.page}>
      <div style={styles.card}>
        <h2 style={styles.title}>Explorador del Sistema de Archivos</h2>
        {tree ? (
          <div>{renderNode(tree)}</div>
        ) : (
          <p style={{ color: "#777" }}>Cargando estructura del sistema de archivos...</p>
        )}
        <button onClick={() => navigate("/discos")} style={styles.backBtn}>
          Volver a Discos
        </button>
      </div>
    </div>
  );
}

const styles = {
  page: {
    display: "flex",
    justifyContent: "center",
    alignItems: "center",
    minHeight: "100vh",
    backgroundColor: "#f4f4f4",
    fontFamily: "Segoe UI, sans-serif",
    padding: "2rem",
  },
  card: {
    backgroundColor: "#ffffff",
    padding: "2rem",
    borderRadius: "12px",
    boxShadow: "0 4px 16px rgba(0,0,0,0.1)",
    width: "100%",
    maxWidth: "800px",
    textAlign: "left",
  },
  title: {
    fontSize: "1.6rem",
    marginBottom: "1rem",
    textAlign: "center",
  },
  backBtn: {
    marginTop: "2rem",
    padding: "0.6rem 1.2rem",
    backgroundColor: "#444",
    color: "white",
    border: "none",
    borderRadius: "6px",
    cursor: "pointer",
    fontWeight: "bold",
  },
};

export default FileViewer;
