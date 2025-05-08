import React, { useEffect, useState } from "react";
import { useLocation, useParams, useNavigate } from "react-router-dom";

function DiskViewer() {
  const { nombre } = useParams();
  const location = useLocation();
  const disk = location.state?.disk;
  const [structure, setStructure] = useState(null);
  const navigate = useNavigate();

  useEffect(() => {
    if (disk?.partition_id) {
      fetch(`http://localhost:8080/filesystem/${disk.partition_id}`)

        .then((res) => res.json())
        .then((data) => setStructure(data))
        .catch(() => setStructure({ error: true }));
    }
  }, [disk]);

  const renderTree = (node, level = 0) => {
    const paddingLeft = 20 * level;
    return (
      <div key={node.name + level + Math.random()} style={{ paddingLeft }}>
        <p>
          {node.type === "folder" ? "📁" : "📄"} {node.name}
        </p>
        {node.children?.map((child) => renderTree(child, level + 1))}
      </div>
    );
  };

  return (
    <div style={styles.container}>
      <h2 style={styles.title}>Explorador del Disco: {nombre}</h2>

      <button
        onClick={() => navigate("/discos")}
        style={styles.backButton}
      >
        ← Volver a Discos
      </button>

      {disk ? (
        <div style={styles.infoBox}>
          <p><strong>Nombre:</strong> {disk.name}</p>
          <p><strong>Ruta:</strong> {disk.path}</p>
          <p><strong>Tamaño:</strong> {disk.size}</p>
          <p><strong>Fit:</strong> {disk.fit}</p>
          <p><strong>Particiones Montadas:</strong> {disk.mounted_partitions.join(", ")}</p>

          <div style={styles.treeBox}>
            {structure ? (
              structure.error ? (
                <p style={{ color: "red" }}>Error al cargar estructura.</p>
              ) : (
                renderTree(structure)
              )
            ) : (
              <p>Cargando estructura...</p>
            )}
          </div>
        </div>
      ) : (
        <p style={{ color: "red" }}>No se encontró información del disco.</p>
      )}
    </div>
  );
}

const styles = {
  container: {
    padding: "2rem",
    fontFamily: "Segoe UI, sans-serif",
  },
  title: {
    fontSize: "1.8rem",
    marginBottom: "1rem",
  },
  backButton: {
    padding: "0.5rem 1rem",
    marginBottom: "1.5rem",
    backgroundColor: "#1976d2",
    color: "white",
    border: "none",
    borderRadius: "6px",
    cursor: "pointer",
    fontWeight: "bold",
  },
  infoBox: {
    backgroundColor: "#f2f2f2",
    padding: "1rem",
    borderRadius: "8px",
  },
  treeBox: {
    marginTop: "2rem",
    padding: "1rem",
    backgroundColor: "#ffffff",
    border: "1px dashed #ccc",
    borderRadius: "6px",
    maxHeight: "500px",
    overflowY: "auto",
  },
};

export default DiskViewer;
