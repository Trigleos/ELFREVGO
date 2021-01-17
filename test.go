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
	section_headers string
	dynamic_strings string
}


func read_file(filename string) (data []byte) {
	data,err := ioutil.ReadFile(filename)
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

func read_sh_table(data []byte, curr_elf ELF) ELF {

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
	
	curr_elf.section_headers = string(data_copy)
	return curr_elf

}

func get_section_by_name(data []byte, curr_elf ELF, searched_section string) [3]int{
	searched_index := strings.Index(curr_elf.section_headers,searched_section)
	var result [3]int
	
	for index:=0;index<curr_elf.number_of_sections;index++ {
		section_header := data[curr_elf.section_table +
				index * curr_elf.size_of_section_headers:
				curr_elf.section_table +
				(index+1) * curr_elf.size_of_section_headers]
				
		string_offset := int(binary.LittleEndian.Uint32(section_header[0:4]))
		if string_offset == searched_index {
			result[0] = int(binary.LittleEndian.Uint64(section_header[24:32]))
			result[1] = int(binary.LittleEndian.Uint64(section_header[32:40]))
			result[2] = int(binary.LittleEndian.Uint64(section_header[56:64]))
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

func read_dyn_str(data []byte, curr_elf ELF) ELF{
	dynstr := get_section_by_name(data,curr_elf, ".dynstr")
	start_offset := dynstr[0]
	length := dynstr[1]
	
	data_copy := data[start_offset:start_offset+length]
	
	for index:=0; index < len(data_copy);index++ {
		if data_copy[index] == 0 {
			data_copy[index] = byte(' ')
		}
	}
	
	curr_elf.dynamic_strings = string(data_copy)
	return curr_elf
}

func get_dyn_function_id_by_name(data []byte, curr_elf ELF, searched_string string) int{
	searched_index := strings.Index(curr_elf.dynamic_strings,searched_string)


	dynsym := get_section_by_name(data, curr_elf, ".dynsym")
	dynsym_section := data[dynsym[0]:dynsym[0]+dynsym[1]]
	
	for index := 0; index<dynsym[1]/dynsym[2]; index++{
		str_index := int(binary.LittleEndian.Uint32(dynsym_section[index*dynsym[2]:index*dynsym[2]+4]))
		if str_index == searched_index {
			return index
		}
		
	}
	return 0
}

func main() {
	
	data := read_file("nopie")
	curr_elf := initialize_ELF(data)
	curr_elf = read_sh_table(data,curr_elf)
	fmt.Println(curr_elf.section_headers)
	
	curr_elf = read_dyn_str(data, curr_elf)
	fmt.Println(curr_elf.dynamic_strings)
	
	fmt.Println(get_section_by_name(data,curr_elf, ".dynsym"))
	
	index := get_dyn_function_id_by_name(data, curr_elf, "printf")
	fmt.Println(index)
}
