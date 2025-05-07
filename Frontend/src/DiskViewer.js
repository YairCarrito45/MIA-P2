import React, { useEffect, useState } from "react";
import { useLocation, useParams } from "react-router-dom";

function DiskViewer() {
  const { nombre } = useParams();
  const location = useLocation();
  const disk = location.state?.disk;
  const [structure, setStructure] = useState(null);

  useEffect(() => {
    if (disk?.partition_id) {
      fetch(`http://localhost:3001/filesystem/${disk.partition_id}`)
        .then((res) => res.json())
        .then((data) => setStructure(data))
        .catch(() => setStructure({ error: true }));
    }
  }, [disk]);

  const renderTree = (node, level = 0) => {
    const paddingLeft = 20 * level;
    return (
      <div key={node.name + level + Math.random()} style={{ paddingLeft }}>
        <p style={{ margin: "4px 0" }}>
          {node.type === "folder" ? "📁" : "📄"} {node.name}
        </p>
        {node.children?.map((child) => renderTree(child, level + 1))}
      </div>
    );
  };

  return (
    <div style={styles.page}>
      <div style={styles.card}>
        <h2 style={styles.title}>Visualizador del Sistema de Archivos</h2>
        <p style={styles.subtitle}>Disco seleccionado: <strong>{nombre}</strong></p>

        {disk ? (
          <>
            <div style={styles.infoGrid}>
              <p><strong>Nombre:</strong> {disk.name}</p>
              <p><strong>Tamaño:</strong> {disk.size}</p>
              <p><strong>Fit:</strong> {disk.fit}</p>
              <p><strong>Particiones Montadas:</strong> {disk.mounted_partitions.join(", ")}</p>
            </div>

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
          </>
        ) : (
          <p style={{ color: "red" }}>No se encontró información del disco.</p>
        )}
      </div>
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
    padding: "2rem",
    fontFamily: "Segoe UI, sans-serif",
  },
  card: {
    backgroundColor: "#ffffff",
    borderRadius: "12px",
    padding: "2rem 3rem",
    boxShadow: "0 4px 12px rgba(0,0,0,0.1)",
    maxWidth: "700px",
    width: "100%",
    textAlign: "center",
  },
  title: {
    fontSize: "1.8rem",
    marginBottom: "0.5rem",
  },
  subtitle: {
    fontSize: "1rem",
    marginBottom: "1.5rem",
    color: "#444",
  },
  infoGrid: {
    textAlign: "left",
    marginBottom: "1.5rem",
    lineHeight: "1.6",
  },
  treeBox: {
    backgroundColor: "#f4f4f4",
    padding: "1rem",
    borderRadius: "6px",
    border: "1px dashed #ccc",
    textAlign: "left",
    maxHeight: "400px",
    overflowY: "auto",
  },
};

export default DiskViewer;
