package main

import (
	"fmt"
	"io/ioutil"
	"encoding/binary"
)


func read_file(filename string) (data []byte) {
	data,err := ioutil.ReadFile("normal")
	if err != nil {
		fmt.Println("File reading error", err)
		return
	}
	
	return
}

func write_file(filename string, data []byte) error {
	err := ioutil.WriteFile(filename,data,0777)
	if err != nil {
		fmt.Println("File writing error",err)
		return err
	}
	
	return err
}

func change_class(data []byte) []byte {
	if data[4] == 2 {
		data[4] = 1
	} else {
		data[4] = 2
	}
	return data
}

func change_endianness(data []byte) []byte {
	if data[5] == 2 {
		data[5] = 1
	} else {
		data[5] = 2
	}
	return data
}

func main() {
	
	data := read_file("normal")
	
	var slice []byte
	slice = data[40:48]
	address := binary.LittleEndian.Uint64(slice)
	
	slice = data[58:60]
	size_sec := binary.LittleEndian.Uint16(slice)
	
	slice = data[60:62]
	num_sec := binary.LittleEndian.Uint16(slice)
	
	slice = data[62:64]
	shstrtab_add := int(binary.LittleEndian.Uint16(slice)) * int(size_sec) + int(address)
	
	fmt.Println(address, size_sec, num_sec, shstrtab_add)
	
	slice = shstrtab[24:32]
	str_tab_off := binary.LittleEndian.Uint64(slice)
	fmt.Println(str_tab_off)
	
	data = change_class(data)
	data = change_endianness(data)
	
	write_file("normal_go",data)
	
}
