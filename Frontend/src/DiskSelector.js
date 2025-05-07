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

  const handleSelect = (disk) => {
    navigate(`/viewer/${disk.name}`, { state: { disk } });
  };

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
              onClick={() => handleSelect(disk)}
              title={`Capacidad: ${disk.size}\nFit: ${disk.fit}\nParticiones: ${disk.mounted_partitions.join(", ") || "Ninguna"}`}
            >
              <img
                src="/disk-icon.png"
                alt="Disco"
                style={styles.diskImage}
              />
              <p style={styles.diskLabel}>{disk.name}</p>
            </div>
          ))}
        </div>
      </div>
  
      {/* 🔙 Botón de volver al menú */}
      <button
        onClick={() => navigate("/")}
        style={{
          marginTop: "2rem",
          padding: "0.6rem 1.2rem",
          backgroundColor: "#444",
          color: "white",
          border: "none",
          borderRadius: "6px",
          cursor: "pointer",
          fontWeight: "bold",
        }}
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
  },
  card: {
    backgroundColor: "#ffffff",
    padding: "2.5rem 3rem",
    borderRadius: "12px",
    boxShadow: "0 4px 12px rgba(0,0,0,0.1)",
    textAlign: "center",
    maxWidth: "700px",
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
    cursor: "pointer",
    padding: "1rem",
    borderRadius: "10px",
    backgroundColor: "#eef7f9",
    width: "100px",
    height: "120px",
    display: "flex",
    flexDirection: "column",
    alignItems: "center",
    justifyContent: "center",
    transition: "transform 0.2s",
    boxShadow: "0 2px 6px rgba(0,0,0,0.1)",
  },
  diskImage: {
    width: "48px",
    marginBottom: "0.5rem",
  },
  diskLabel: {
    fontSize: "0.9rem",
    fontWeight: "bold",
  },
};

export default DiskSelector;
