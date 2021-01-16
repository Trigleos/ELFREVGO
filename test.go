package main

import (
	"fmt"
	"io/ioutil"
	"encoding/binary"
	"strings"
)

type ELF struct {
	section_table int
	number_of_sections int
	size_of_section_headers int
}


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

func read_sh_table(data []byte, curr_elf ELF) string {

	slice := data[62:64]
	sh_table_header_addr := int(binary.LittleEndian.Uint16(slice)) * curr_elf.size_of_section_headers + curr_elf.section_table
	
	sh_table_header := data[sh_table_header_addr:sh_table_header_addr+curr_elf.size_of_section_headers]
	
	slice = sh_table_header[24:32]
	sh_table_addr := int(binary.LittleEndian.Uint64(slice))


	curr_strings := -1
	index := sh_table_addr
	var curr_byte int
	
	for curr_strings < curr_elf.number_of_sections {
		curr_byte = int(data[index])
		if curr_byte == 0 {
			curr_strings++
		}
		index++
	}
	
	data_copy := data[sh_table_addr:index]
	
	for index:=0; index < len(data_copy);index++ {
		if data_copy[index] == 0 {
			data_copy[index] = byte(' ')
		}
	}
	
	return string(data_copy)

}

func get_section_by_name(data []byte, curr_elf ELF, section_headers string, searched_section string) [2]int{
	searched_index := strings.Index(section_headers,searched_section)
	var result [2]int
	
	for index:=0;index<curr_elf.number_of_sections;index++ {
		section_header := data[curr_elf.section_table +
				index * curr_elf.size_of_section_headers:
				curr_elf.section_table +
				(index+1) * curr_elf.size_of_section_headers]
				
		string_offset := int(binary.LittleEndian.Uint32(section_header[0:4]))
		if string_offset == searched_index {
			result[0] = int(binary.LittleEndian.Uint64(section_header[24:32]))
			result[1] = int(binary.LittleEndian.Uint64(section_header[32:40]))
		}
	}
	
	return result
}

func initialize_ELF(data []byte) ELF {
	var curr_elf = ELF{}
	
	slice := data[40:48]
	curr_elf.section_table = int(binary.LittleEndian.Uint64(slice))
	
	slice = data[58:60]
	curr_elf.size_of_section_headers = int(binary.LittleEndian.Uint16(slice))
	
	slice = data[60:62]
	curr_elf.number_of_sections = int(binary.LittleEndian.Uint16(slice))
	
	return curr_elf
}

func main() {
	
	data := read_file("normal")
	curr_elf := initialize_ELF(data)
	section_headers := read_sh_table(data,curr_elf)
	fmt.Println(section_headers)
	fmt.Println(get_section_by_name(data,curr_elf, section_headers, ".data"))	
}
