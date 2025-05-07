import React, { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";

function DiskSelector() {
  const [disks, setDisks] = useState([]);
  const navigate = useNavigate();

  useEffect(() => {
    fetch("http://localhost:3001/disks")
      .then((res) => res.json())
      .then((data) => setDisks(data))
      .catch((err) => console.error("Error al cargar discos:", err));
  }, []);

  return (
    <div style={styles.page}>
      <div style={styles.card}>
        <h2 style={styles.title}>Visualizador del Sistema de Archivos</h2>
        <p style={styles.subtitle}>Seleccione el disco que desea visualizar:</p>
        <div style={styles.grid}>
          {disks.map((disk) => (
            <div
              key={disk.name}
              style={styles.diskCard}
              onClick={() => {
                localStorage.setItem("selectedDiskPath", disk.path);
                navigate(`/partitions/${encodeURIComponent(disk.name)}`);
              }}
              
              
            >
              <img src="/disk-icon.png" alt="Disco" style={styles.diskImage} />
              <p style={styles.diskLabel}>{disk.name}</p>
              <div style={styles.diskInfo}>
                <p><strong>Ruta:</strong> {disk.path}</p>
                <p><strong>Capacidad:</strong> {disk.size}</p>
                <p><strong>Fit:</strong> {disk.fit}</p>
                <p><strong>Particiones:</strong> {disk.mounted_partitions.join(", ") || "Ninguna"}</p>
              </div>
            </div>
          ))}
        </div>
      </div>

      <button
        onClick={() => navigate("/")}
        style={styles.backBtn}
      >
        Volver al Menú Principal
      </button>
    </div>
  );
}

const styles = {
  page: {
    backgroundColor: "#f8f8f8",
    minHeight: "100vh",
    display: "flex",
    justifyContent: "center",
    alignItems: "center",
    fontFamily: "Segoe UI, sans-serif",
    padding: "2rem",
    flexDirection: "column",
  },
  card: {
    backgroundColor: "#ffffff",
    padding: "2.5rem 3rem",
    borderRadius: "12px",
    boxShadow: "0 4px 12px rgba(0,0,0,0.1)",
    textAlign: "center",
    maxWidth: "900px",
    width: "100%",
  },
  title: {
    fontSize: "1.8rem",
    marginBottom: "0.5rem",
  },
  subtitle: {
    fontSize: "1rem",
    marginBottom: "2rem",
    color: "#555",
  },
  grid: {
    display: "flex",
    justifyContent: "center",
    gap: "2rem",
    flexWrap: "wrap",
  },
  diskCard: {
    padding: "1rem",
    borderRadius: "10px",
    backgroundColor: "#eef7f9",
    width: "220px",
    boxShadow: "0 2px 6px rgba(0,0,0,0.1)",
    display: "flex",
    flexDirection: "column",
    alignItems: "center",
    transition: "transform 0.2s",
    cursor: "pointer",
  },
  diskImage: {
    width: "48px",
    marginBottom: "0.5rem",
  },
  diskLabel: {
    fontSize: "0.95rem",
    fontWeight: "bold",
    marginBottom: "0.4rem",
  },
  diskInfo: {
    fontSize: "0.75rem",
    textAlign: "left",
    lineHeight: "1.2rem",
    width: "100%",
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

export default DiskSelector;
