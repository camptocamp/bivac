package providers

/*
func isMySQL(volumePath string) (err error, detected bool) {
	files, err = listFiles(volumePath)
	if err != nil {
		return
	}

	if mySQLConfigFileExists(files) {
		detected = true
	}
	return
}

func mySQLConfigFileExists(files []string) bool {
	dbConfigFile := "my.cnf"

	for _, f := range files {
		if strings.Contains(f, dbConfigFile) {
			spStr := strings.Split(f, "/")
			if spStr[len(spStr)-1] == dbConfigFile {
				return true
			}
		}
	}
	return false
}
*/
