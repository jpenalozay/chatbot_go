package initializers

import (
	"chatbot/logger"
	"chatbot/models"
	"fmt"

	"gorm.io/gorm"
)

// Migrate realiza la migración de los modelos a la base de datos
func Migrate() error {
	logger.Log.Info("Iniciando migración de la base de datos...")

	// Conectar a la base de datos si no está conectado
	if err := InitPostgres(); err != nil {
		logger.Log.Errorf("Error al inicializar PostgreSQL: %v", err)
		return err
	}

	// Realiza la migración de los modelos
	err := DB.AutoMigrate(&models.User{}, &models.Role{}, &models.UsuarioChat{}, &models.Hilo{}, &models.Mensaje{}, &models.Interes{}, &models.CatalogoInteres{})
	if err != nil {
		logger.Log.Errorf("Error al migrar la base de datos: %v", err)
		return fmt.Errorf("error al migrar la base de datos: %v", err)
	}

	logger.Log.Info("Migración de la base de datos completada.")

	// Crear roles por defecto si no existen
	createRoleIfNotExists(DB, models.AdminRole)
	createRoleIfNotExists(DB, models.UserRole)

	// Crear usuario de prueba si no existe
	createTestUserIfNotExists(DB, "jlpy", "jlpy")

	// Poblar la tabla CatalogoInteres si está vacía
	if err := populateCatalogoInteres(DB); err != nil {
		logger.Log.Errorf("Error al poblar CatalogoInteres: %v", err)
	}

	logger.Log.Info("Roles, usuario de prueba y catálogo de intereses creados exitosamente.")
	return nil
}

// Función para poblar la tabla CatalogoInteres
func populateCatalogoInteres(db *gorm.DB) error {
	logger.Log.Info("Poblando tabla CatalogoInteres...")

	var count int64
	db.Model(&models.CatalogoInteres{}).Count(&count)
	if count > 0 {
		logger.Log.Info("La tabla CatalogoInteres ya contiene datos.")
		return nil
	}

	intereses := []models.CatalogoInteres{
		{Codigo: "1001", Descripcion: "Descripcion de la Carrera de Medicina Humana"},
		{Codigo: "1002", Descripcion: "Campus y Modalidades de Medicina Humana"},
		{Codigo: "1003", Descripcion: "Costos de la Carrera de Medicina Humana"},
		{Codigo: "1004", Descripcion: "Malla Curricular de Medicina Humana"},
		{Codigo: "1005", Descripcion: "Descripcion de la Carrera de Psicologia"},
		{Codigo: "1006", Descripcion: "Campus y Modalidades de Psicologia"},
		{Codigo: "1007", Descripcion: "Costos de la Carrera de Psicologia"},
		{Codigo: "1008", Descripcion: "Malla Curricular de Psicologia"},
		{Codigo: "1009", Descripcion: "Descripcion de la Carrera de Terapia Fisica y Rehabilitacion"},
		{Codigo: "1010", Descripcion: "Campus y Modalidades de Terapia Fisica y Rehabilitacion"},
		{Codigo: "1011", Descripcion: "Costos de la Carrera de Terapia Fisica y Rehabilitacion"},
		{Codigo: "1012", Descripcion: "Malla Curricular de Terapia Fisica y Rehabilitacion"},
		{Codigo: "1013", Descripcion: "Descripcion de la Carrera de Enfermeria"},
		{Codigo: "1014", Descripcion: "Campus y Modalidades de Enfermeria"},
		{Codigo: "1015", Descripcion: "Costos de la Carrera de Enfermeria"},
		{Codigo: "1016", Descripcion: "Malla Curricular de Enfermeria"},
		{Codigo: "1017", Descripcion: "Descripcion de la Carrera de Tecnologia Medica"},
		{Codigo: "1018", Descripcion: "Campus y Modalidades de Tecnologia Medica"},
		{Codigo: "1019", Descripcion: "Costos de la Carrera de Tecnologia Medica"},
		{Codigo: "1020", Descripcion: "Malla Curricular de Tecnologia Medica"},
		{Codigo: "1021", Descripcion: "Descripcion de la Carrera de Ingenieria Civil"},
		{Codigo: "1022", Descripcion: "Campus y Modalidades de Ingenieria Civil"},
		{Codigo: "1023", Descripcion: "Costos de la Carrera de Ingenieria Civil"},
		{Codigo: "1024", Descripcion: "Malla Curricular de Ingenieria Civil"},
		{Codigo: "1025", Descripcion: "Descripcion de la Carrera de Ingenieria de Sistemas"},
		{Codigo: "1026", Descripcion: "Campus y Modalidades de Ingenieria de Sistemas"},
		{Codigo: "1027", Descripcion: "Costos de la Carrera de Ingenieria de Sistemas"},
		{Codigo: "1028", Descripcion: "Malla Curricular de Ingenieria de Sistemas"},
		{Codigo: "1029", Descripcion: "Descripcion de la Carrera de Ingenieria Industrial"},
		{Codigo: "1030", Descripcion: "Campus y Modalidades de Ingenieria Industrial"},
		{Codigo: "1031", Descripcion: "Costos de la Carrera de Ingenieria Industrial"},
		{Codigo: "1032", Descripcion: "Malla Curricular de Ingenieria Industrial"},
		{Codigo: "1033", Descripcion: "Descripcion de la Carrera de Ingenieria Ambiental"},
		{Codigo: "1034", Descripcion: "Campus y Modalidades de Ingenieria Ambiental"},
		{Codigo: "1035", Descripcion: "Costos de la Carrera de Ingenieria Ambiental"},
		{Codigo: "1036", Descripcion: "Malla Curricular de Ingenieria Ambiental"},
		{Codigo: "1037", Descripcion: "Descripcion de la Carrera de Ingenieria Mecanica"},
		{Codigo: "1038", Descripcion: "Campus y Modalidades de Ingenieria Mecanica"},
		{Codigo: "1039", Descripcion: "Costos de la Carrera de Ingenieria Mecanica"},
		{Codigo: "1040", Descripcion: "Malla Curricular de Ingenieria Mecanica"},
		{Codigo: "1041", Descripcion: "Descripcion de la Carrera de Ingenieria Electronica"},
		{Codigo: "1042", Descripcion: "Campus y Modalidades de Ingenieria Electronica"},
		{Codigo: "1043", Descripcion: "Costos de la Carrera de Ingenieria Electronica"},
		{Codigo: "1044", Descripcion: "Malla Curricular de Ingenieria Electronica"},
		{Codigo: "1045", Descripcion: "Descripcion de la Carrera de Administracion"},
		{Codigo: "1046", Descripcion: "Campus y Modalidades de Administracion"},
		{Codigo: "1047", Descripcion: "Costos de la Carrera de Administracion"},
		{Codigo: "1048", Descripcion: "Malla Curricular de Administracion"},
		{Codigo: "1049", Descripcion: "Descripcion de la Carrera de Contabilidad"},
		{Codigo: "1050", Descripcion: "Campus y Modalidades de Contabilidad"},
		{Codigo: "1051", Descripcion: "Costos de la Carrera de Contabilidad"},
		{Codigo: "1052", Descripcion: "Malla Curricular de Contabilidad"},
		{Codigo: "1053", Descripcion: "Descripcion de la Carrera de Finanzas y Negocios Internacionales"},
		{Codigo: "1054", Descripcion: "Campus y Modalidades de Finanzas y Negocios Internacionales"},
		{Codigo: "1055", Descripcion: "Costos de la Carrera de Finanzas y Negocios Internacionales"},
		{Codigo: "1056", Descripcion: "Malla Curricular de Finanzas y Negocios Internacionales"},
		{Codigo: "1057", Descripcion: "Descripcion de la Carrera de Marketing y Gestion Comercial"},
		{Codigo: "1058", Descripcion: "Campus y Modalidades de Marketing y Gestion Comercial"},
		{Codigo: "1059", Descripcion: "Costos de la Carrera de Marketing y Gestion Comercial"},
		{Codigo: "1060", Descripcion: "Malla Curricular de Marketing y Gestion Comercial"},
		{Codigo: "1061", Descripcion: "Descripcion de la Carrera de Economia"},
		{Codigo: "1062", Descripcion: "Campus y Modalidades de Economia"},
		{Codigo: "1063", Descripcion: "Costos de la Carrera de Economia"},
		{Codigo: "1064", Descripcion: "Malla Curricular de Economia"},
		{Codigo: "1065", Descripcion: "Descripcion de la Carrera de Derecho y Ciencias Politicas"},
		{Codigo: "1066", Descripcion: "Campus y Modalidades de Derecho y Ciencias Politicas"},
		{Codigo: "1067", Descripcion: "Costos de la Carrera de Derecho y Ciencias Politicas"},
		{Codigo: "1068", Descripcion: "Malla Curricular de Derecho y Ciencias Politicas"},
		{Codigo: "1069", Descripcion: "Descripcion de la Carrera de Educacion Inicial"},
		{Codigo: "1070", Descripcion: "Campus y Modalidades de Educacion Inicial"},
		{Codigo: "1071", Descripcion: "Costos de la Carrera de Educacion Inicial"},
		{Codigo: "1072", Descripcion: "Malla Curricular de Educacion Inicial"},
		{Codigo: "1073", Descripcion: "Descripcion de la Carrera de Educacion Primaria"},
		{Codigo: "1074", Descripcion: "Campus y Modalidades de Educacion Primaria"},
		{Codigo: "1075", Descripcion: "Costos de la Carrera de Educacion Primaria"},
		{Codigo: "1076", Descripcion: "Malla Curricular de Educacion Primaria"},
		{Codigo: "1077", Descripcion: "Descripcion de la Carrera de Educacion Secundaria"},
		{Codigo: "1078", Descripcion: "Campus y Modalidades de Educacion Secundaria"},
		{Codigo: "1079", Descripcion: "Costos de la Carrera de Educacion Secundaria"},
		{Codigo: "1080", Descripcion: "Malla Curricular de Educacion Secundaria"},
		{Codigo: "1081", Descripcion: "Descripcion de la Carrera de Arquitectura"},
		{Codigo: "1082", Descripcion: "Campus y Modalidades de Arquitectura"},
		{Codigo: "1083", Descripcion: "Costos de la Carrera de Arquitectura"},
		{Codigo: "1084", Descripcion: "Malla Curricular de Arquitectura"},
		{Codigo: "1085", Descripcion: "Descripcion de la Carrera de Diseno de Interiores"},
		{Codigo: "1086", Descripcion: "Campus y Modalidades de Diseno de Interiores"},
		{Codigo: "1087", Descripcion: "Costos de la Carrera de Diseno de Interiores"},
		{Codigo: "1088", Descripcion: "Malla Curricular de Diseno de Interiores"},
		{Codigo: "1089", Descripcion: "Descripcion de la Carrera de Trabajo Social"},
		{Codigo: "1090", Descripcion: "Campus y Modalidades de Trabajo Social"},
		{Codigo: "1091", Descripcion: "Costos de la Carrera de Trabajo Social"},
		{Codigo: "1092", Descripcion: "Malla Curricular de Trabajo Social"},
		{Codigo: "1093", Descripcion: "Descripcion de la Carrera de Ciencias de la Comunicacion"},
		{Codigo: "1094", Descripcion: "Campus y Modalidades de Ciencias de la Comunicacion"},
		{Codigo: "1095", Descripcion: "Costos de la Carrera de Ciencias de la Comunicacion"},
		{Codigo: "1096", Descripcion: "Malla Curricular de Ciencias de la Comunicacion"},
		{Codigo: "2001", Descripcion: "Proceso de Admision"},
		{Codigo: "2002", Descripcion: "Fecha de Examen de Admision"},
		{Codigo: "2003", Descripcion: "Costo del Examen de Admision"},
		{Codigo: "2004", Descripcion: "Requisitos para el Examen de Admision"},
		{Codigo: "3001", Descripcion: "Informacion General sobre la Universidad Continental"},
		{Codigo: "3002", Descripcion: "Por que estudiar en la Universidad Continental"},
		{Codigo: "3003", Descripcion: "Beneficios de estudiar en la Universidad Continental"},
		{Codigo: "3004", Descripcion: "Infraestructura y Servicios de la Universidad Continental"},
		{Codigo: "3005", Descripcion: "Convenios y Alianzas de la Universidad Continental"},
	}

	for _, interes := range intereses {
		if err := db.Create(&interes).Error; err != nil {
			logger.Log.Errorf("Error al crear interés %s: %v", interes.Codigo, err)
			return fmt.Errorf("error al crear interés %s: %v", interes.Codigo, err)
		}
	}

	logger.Log.Info("Tabla CatalogoInteres poblada exitosamente.")
	return nil
}

// createRoleIfNotExists crea un rol en la base de datos si no existe
func createRoleIfNotExists(db *gorm.DB, roleName string) {
	var role models.Role
	if err := db.Where("name = ?", roleName).First(&role).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			if err := db.Create(&models.Role{Name: roleName}).Error; err != nil {
				logger.Log.Fatalf("error al crear el rol %s: %v", roleName, err)
			}
			logger.Log.Infof("Rol %s creado exitosamente.", roleName)
		} else {
			logger.Log.Fatalf("error al consultar el rol %s: %v", roleName, err)
		}
	} else {
		logger.Log.Infof("El rol %s ya existe.", roleName)
	}
}

// createTestUserIfNotExists crea un usuario de prueba en la base de datos si no existe
func createTestUserIfNotExists(db *gorm.DB, username, password string) {
	var user models.User
	if err := db.Where("username = ?", username).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			user = models.User{Username: username}
			if err := user.SetPassword(password); err != nil {
				logger.Log.Fatalf("error al establecer la contraseña para el usuario de prueba: %v", err)
			}
			if err := db.Create(&user).Error; err != nil {
				logger.Log.Fatalf("error al crear el usuario de prueba: %v", err)
			}
			assignRoleToUser(db, &user, models.UserRole)
			logger.Log.Infof("Usuario de prueba %s creado exitosamente.", username)
		} else {
			logger.Log.Fatalf("error al consultar el usuario %s: %v", username, err)
		}
	} else {
		logger.Log.Infof("El usuario de prueba %s ya existe.", username)
	}
}

// assignRoleToUser asigna un rol a un usuario
func assignRoleToUser(db *gorm.DB, user *models.User, roleName string) {
	var role models.Role
	if err := db.Where("name = ?", roleName).First(&role).Error; err != nil {
		logger.Log.Fatalf("error al encontrar el rol %s: %v", roleName, err)
	}
	if err := db.Model(user).Association("Roles").Append(&role); err != nil {
		logger.Log.Fatalf("error al asignar el rol %s al usuario %s: %v", roleName, user.Username, err)
	}
	logger.Log.Infof("Rol %s asignado al usuario %s exitosamente.", roleName, user.Username)
}
