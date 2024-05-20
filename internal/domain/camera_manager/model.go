package camera_manager

import "monitoring-system/internal/domain/camera"

func (cm *cameraManager) loadCamerasFromDB() error {
	query := `CREATE TABLE IF NOT EXISTS cameras (
		id INTEGER PRIMARY KEY,
		name TEXT,
		status TEXT
	);`
	_, err := cm.db.Exec(query)
	if err != nil {
		return err
	}

	rows, err := cm.db.Query("SELECT id, name, status FROM cameras")
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var cam Camera
		var status string
		if err := rows.Scan(&cam.Id, &cam.Name, &status); err != nil {
			return err
		}
		cam.Status = Status(status)
		cam.Camera = camera.NewWebcam(cm.ctx, cam.Id, cm.logger)
		cm.cameras[cam.Id] = cam
	}

	for i := 0; i < cm.cameraConfig.MaxCameraCount; i++ {
		if _, exists := cm.cameras[i]; !exists {
			cm.cameras[i] = Camera{
				Id:     i,
				Name:   "",
				Status: Disconnected,
				Camera: camera.NewWebcam(cm.ctx, i, cm.logger),
			}
		}
	}

	return nil
}

func (cm *cameraManager) saveCameraToDB(cam Camera) error {
	query := `INSERT INTO cameras (id, name, status) VALUES (?, ?, ?)
	          ON CONFLICT(id) DO UPDATE SET name=excluded.name, status=excluded.status;`
	_, err := cm.db.Exec(query, cam.Id, cam.Name, string(cam.Status))
	return err
}
