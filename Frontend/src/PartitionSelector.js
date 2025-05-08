import React, { useEffect, useState } from "react";
import { useNavigate, useLocation } from "react-router-dom";

function PartitionSelector() {
  const location = useLocation();
  const navigate = useNavigate();

  // ✅ Soporte de fallback con localStorage
  const diskPath = location.state?.path || localStorage.getItem("selectedDiskPath");

  const [partitions, setPartitions] = useState([]);

  useEffect(() => {
    if (diskPath) {
      fetch(`http://localhost:8080/api/partitions?path=${encodeURIComponent(diskPath)}`)
        .then((res) => res.json())
        .then((data) => setPartitions(Array.isArray(data) ? data : []))
        .catch(() => setPartitions([]));
    }
  }, [diskPath]);

  const handleSelectPartition = (partition) => {
    navigate(`/viewer/${partition.id}`, {
      state: {
        disk: {
          name: partition.name,
          partition_id: partition.id,
          path: diskPath,
          size: partition.size,
          fit: partition.fit,
          status: partition.status,
        },
      },
    });
  };

  return (
    <div style={styles.page}>
      <div style={styles.card}>
        <h2 style={styles.title}>Visualizador del Sistema de Archivos</h2>
        <p style={styles.subtitle}>Seleccione la partición que desea visualizar:</p>

        <div style={styles.grid}>
          {partitions.length > 0 ? (
            partitions.map((p) => (
              <div
                key={p.id}
                style={styles.partitionCard}
                onClick={() => handleSelectPartition(p)}
              >
                <img src="/particion-icon.png" alt="Partición" style={{ width: "40px", marginBottom: "8px" }} />
                <p><strong>{p.name}</strong></p>
                <p><strong>Tamaño:</strong> {p.size}</p>
                <p><strong>Fit:</strong> {p.fit}</p>
                <p><strong>Estado:</strong> {p.status}</p>
              </div>
            ))
          ) : (
            <p style={{ color: "#777" }}>No se encontraron particiones activas.</p>
          )}
        </div>

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
  },
  card: {
    backgroundColor: "#ffffff",
    padding: "2.5rem",
    borderRadius: "12px",
    boxShadow: "0 4px 16px rgba(0,0,0,0.1)",
    width: "90%",
    maxWidth: "800px",
    textAlign: "center",
  },
  title: {
    fontSize: "1.8rem",
    marginBottom: "0.5rem",
  },
  subtitle: {
    marginBottom: "1.5rem",
    color: "#444",
  },
  grid: {
    display: "flex",
    justifyContent: "center",
    flexWrap: "wrap",
    gap: "1.5rem",
  },
  partitionCard: {
    backgroundColor: "#eef7f9",
    borderRadius: "10px",
    padding: "1rem",
    cursor: "pointer",
    width: "160px",
    boxShadow: "0 3px 8px rgba(0,0,0,0.1)",
    transition: "transform 0.2s",
  },
  img: {
    width: "48px",
    marginBottom: "0.5rem",
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

export default PartitionSelector;
