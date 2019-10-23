package clang

//func IsCleanFile(invocation *TidyInvocation) (bool, error) {
//	fmt.Println("##### IS CLEAN #####")
//
//	if invocation == nil {
//		fmt.Println("IS CLEAN - No invocation")
//		return false, errors.New("Null clang tidy invoke - internal error")
//	}
//
//	// lookup the info on the file
//	info, err := os.Stat(invocation.ExportFile)
//	if err != nil {
//		fmt.Println("IS CLEAN: ", err)
//		return false, err
//	}
//
//	fmt.Println("IS CLEAN SIZE: ", info.Size())
//
//	return info.Size() == 0, nil
//}
