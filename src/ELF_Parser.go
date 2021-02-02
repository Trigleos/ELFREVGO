package main

import (
	"fmt"
	"io/ioutil"
	"encoding/binary"
	"strings"
	"strconv"
)

type ELF struct {  //ELF struct that is extensively used throughout this script
	bit64 bool
	section_table int
	number_of_sections int
	size_of_section_headers int
	section_headers string
	dynamic_strings string
	string_table string
}


func read_file(filename string) (data []byte) {  //read in bytestream from file
	data,err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Println("File reading error", err)
		return
	}
	
	return
}

func write_file(filename string, data []byte) error {  //write byte stream to a file
	err := ioutil.WriteFile(filename,data,0777)
	if err != nil {
		fmt.Println("File writing error",err)
		return err
	}
	
	return err
}

func change_class(data []byte) []byte {  //change class of file (64 bit or 32 bit) by switching a value in the ELF header
	if data[4] == 2 {
		data[4] = 1
	} else {
		data[4] = 2
	}
	return data
}

func change_endianness(data []byte) []byte {  //change endianness of file by switching a value in the ELF header
	if data[5] == 2 {
		data[5] = 1
	} else {
		data[5] = 2
	}
	return data
}

func write_data(hex_string string, data []byte, addr int, size int) []byte {  //write certain hex string to a specific address in little endian and return the changed byte stream

	for index := 0; index < size; index++ {
		data[addr+index] = 0
	}
	
	hex_data := strings.Replace(hex_string, "0x", "", -1)
	hex_data = strings.Replace(hex_data, "0X", "", -1) 
	

	if len(hex_data) % 2 != 0 {
		hex_data = "0" + hex_data
	}
	length := len(hex_data)/2
	for index := length-1; index >= 0 ; index-- {
		current_int,err := strconv.ParseInt(hex_data[index*2:index*2+2],16,size*8)
		if err == nil {
			data[addr+(length-1-index)] = byte(current_int)
		}
	}
	return data

}

func read_sh_table(data []byte, curr_elf ELF) ELF {  //parses the section header string table that contains names of different sections

	var sh_table_header_addr int
	var sh_table_addr int
	slice := data[0:1]
	
	if curr_elf.bit64 {
		slice = data[62:64]  //This is the index of the section header string table in the section table
		sh_table_header_addr = int(binary.LittleEndian.Uint16(slice)) * curr_elf.size_of_section_headers + curr_elf.section_table  //To get the address of the section header for the section header string table, mutliply the size of the headers by the index of the section header string table and add the initial offset at which the section header table starts
	} else {  //Same process for 32 bit
		slice = data[50:52]
		sh_table_header_addr = int(binary.LittleEndian.Uint16(slice)) * curr_elf.size_of_section_headers + curr_elf.section_table
	}
		
	sh_table_header := data[sh_table_header_addr:sh_table_header_addr+curr_elf.size_of_section_headers]  //gets the entire section header for the section header string table
	
	if curr_elf.bit64 {
		slice = sh_table_header[24:32]  //gets the physical offset for the section header string table
		sh_table_addr = int(binary.LittleEndian.Uint64(slice))
	} else {  //same process for 32 bit
		slice = sh_table_header[16:20]
		sh_table_addr = int(binary.LittleEndian.Uint32(slice))
	}

	//the section header string table is a simple array of strings separated by Null bytes

	curr_strings := -1  
	index := sh_table_addr
	var curr_byte int
	
	for curr_strings < curr_elf.number_of_sections {  //scan section header string table until you have all section header names, sets up index as end of section header string table
		curr_byte = int(data[index])
		if curr_byte == 0 {
			curr_strings++
		}
		index++
	}
	
	var data_copy []byte
	data_copy = make([]byte,index-sh_table_addr)
	copy(data_copy,data[sh_table_addr:index]) //copy the entire string table
	
	for index:=0; index < len(data_copy);index++ {
		if data_copy[index] == 0 {
			data_copy[index] = byte(' ')	//replace null bytes by spaces
		}
	}
	
	curr_elf.section_headers = string(data_copy) //add the whole thing as a single big string to the ELF struct
	return curr_elf

}

func get_section_by_name(data []byte, curr_elf ELF, searched_section string) [4]int{ //uses section header string table and name of searched section to return address of a specific section
	searched_index := strings.Index(curr_elf.section_headers,searched_section)  //In the section headers, sections refer to their names with an index into the section header string table
	var result [4]int
	
	//function returns several key information on searched section:
	//1: physical offset of section in the file
	//2: size of the section in bytes
	//3: the size in bytes of each entry, for sections that contain fixed-size entries, important for sections that are tables
	//4: The virtual address of a section in memory, this is important to later convert virtual addresses to physical offsets
	
	for index:=0;index<curr_elf.number_of_sections;index++ {  //gets a whole section header using the initial offset of the section header table and the size for each section header
		section_header := data[curr_elf.section_table +
				index * curr_elf.size_of_section_headers:
				curr_elf.section_table +
				(index+1) * curr_elf.size_of_section_headers]
				
		string_offset := int(binary.LittleEndian.Uint32(section_header[0:4]))  //this refers to the name of a section, as mentioned above
		if string_offset == searched_index {
			if curr_elf.bit64 {
				result[0] = int(binary.LittleEndian.Uint64(section_header[24:32]))  //gets physical offset
				result[1] = int(binary.LittleEndian.Uint64(section_header[32:40]))  //gets size of section
				result[2] = int(binary.LittleEndian.Uint64(section_header[56:64]))  //gets size of entry
				result[3] = int(binary.LittleEndian.Uint64(section_header[16:24]))  //gets virtual address
			} else {  //same process for 32 bits
				result[0] = int(binary.LittleEndian.Uint32(section_header[16:20]))
				result[1] = int(binary.LittleEndian.Uint32(section_header[20:24]))
				result[2] = int(binary.LittleEndian.Uint32(section_header[36:40]))
				result[3] = int(binary.LittleEndian.Uint32(section_header[12:16]))
			}
		}
	}
	
	return result  //returns the information on the section
}

func initialize_ELF(data []byte) ELF {  //initializes ELF struct by parsing ELF header
	var curr_elf = ELF{}
	curr_elf.bit64 = check_for_64bit(data)  //checks if ELF is 64 or 32 bit
	
	if curr_elf.bit64 {
		slice := data[40:48]
		curr_elf.section_table = int(binary.LittleEndian.Uint64(slice))  //gets offset of section table
		
		slice = data[58:60]
		curr_elf.size_of_section_headers = int(binary.LittleEndian.Uint16(slice)) //gets size of each section header in the section table
	
		slice = data[60:62]
		curr_elf.number_of_sections = int(binary.LittleEndian.Uint16(slice))  //gets number of sections in the section table
		
	} else {
		slice := data[32:36]
		curr_elf.section_table = int(binary.LittleEndian.Uint32(slice))  //gets offset of section table
		
		slice = data[46:48]
		curr_elf.size_of_section_headers = int(binary.LittleEndian.Uint16(slice))  //gets size of each section header in the section table
	
		slice = data[48:50]
		curr_elf.number_of_sections = int(binary.LittleEndian.Uint16(slice))  //gets number of sections in the section table
	}
	
	
	curr_elf = read_sh_table(data, curr_elf)  //parses section header string table that contains names of sections
	
	curr_elf = read_dyn_str(data, curr_elf)  //parses dynamic string table that contains names for dynamic functions
	
	curr_elf = read_str_tab(data, curr_elf) //parses string table that contains names for normal functions
	
	return curr_elf
}

func read_dyn_str(data []byte, curr_elf ELF) ELF{  //reads the dynamic string table, a table that contains the names for dynamic functions such as external library functions
	dynstr := get_section_by_name(data,curr_elf, ".dynstr")  //this is where the names are stored
	start_offset := dynstr[0] //start of the section
	length := dynstr[1] //size of the section
	
	var data_copy []byte
	data_copy = make([]byte,length)
	copy(data_copy,data[start_offset:start_offset+length])  //takes a copy from the whole section
	
	//the dynstr section is basically a long array containing zero-terminated strings that represent the names for the dynamic symbols
	
	for index:=0; index < len(data_copy);index++ {  //loop over the section and replace null bytes with spaces
		if data_copy[index] == 0 {
			data_copy[index] = byte(' ')
		}
	}
	
	curr_elf.dynamic_strings = string(data_copy)  //convert the whole section into one big string and save it in the ELF struct
	return curr_elf
}

func read_str_tab(data []byte, curr_elf ELF) ELF{  //reads the string table, a table that contains the names for every symbol such as function names
	strtab := get_section_by_name(data,curr_elf, ".strtab")  //this is where the names are stored
	if strtab[0] != 0 {
		start_offset := strtab[0] //start of the section
		length := strtab[1] //size of the section
		
		var data_copy []byte
		data_copy = make([]byte,length)
		copy(data_copy,data[start_offset:start_offset+length])  //takes a copy from the whole section
		
		//the strtab section is basically a long array containing zero-terminated strings that represent the names for the symbols
		
		for index:=0; index < len(data_copy);index++ {  //loop over the section and replace null bytes with spaces
			if data_copy[index] == 0 {
				data_copy[index] = byte(' ')
			}
		}
		
		curr_elf.string_table = string(data_copy)  //convert the whole section into one big string and save it in the ELF struct
		return curr_elf
	} else {
		return curr_elf
	}
}

func get_dyn_function_id_by_name(data []byte, curr_elf ELF, searched_string string) int{  //returns ID for dynamic function. This ID is used internally by the runtime linker to resolve dynamic function addresses and we need it to get the GOT address
	searched_index := strings.Index(curr_elf.dynamic_strings,searched_string)  //The dynsym table again refers to the name of each dynamic symbol using an offset into the dynstr table which we parsed earlier
	
	dynsym := get_section_by_name(data, curr_elf, ".dynsym") //dynsym is the table that contains all the dynamic symbols
	dynsym_section := data[dynsym[0]:dynsym[0]+dynsym[1]]  //get the entire section
	
	for index := 0; index<dynsym[1]/dynsym[2]; index++{  //we can easily loop over this section because we know the size for each entry in the dynsym table
		str_index := int(binary.LittleEndian.Uint32(dynsym_section[index*dynsym[2]:index*dynsym[2]+4]))  //This is an index into the dynstr table
		if str_index == searched_index {
			return index  //the dynamic symbols are simply numbered from top to bottom so we can simply return the index
		}
		
	}
	return 0
}

func get_dyn_addr_by_name(data []byte, curr_elf ELF, searched_function string) int{  //returns the virtual address for a dynamic function
	id := get_dyn_function_id_by_name(data,curr_elf,searched_function)  //gets the ID for the specific function we're searching for
	if curr_elf.bit64 {  
		rela_plt := get_section_by_name(data,curr_elf,".rela.plt")  //The rela.plt section contains the relocation information that is used by the runtime linker to determine where he needs to put the addresses for the dynamic library functions. More specifically it contains addresses in the GOT, which we want to overwrite.
		rela_plt_section := data[rela_plt[0]:rela_plt[0]+rela_plt[1]] //get entire section
		for index:=0;index < rela_plt[1]/rela_plt[2];index++ {  //loop over this section 
			plt_entry := rela_plt_section[index*rela_plt[2]:(index+1)*rela_plt[2]]  //get one entry
			info := int(binary.LittleEndian.Uint64(plt_entry[8:16])) >> 32  //info is the second field in each rela.plt entry. It links each entry to its name and contains other information. To get the function ID, we need to shift the value to the left by 32
			if info == id {
				return int(binary.LittleEndian.Uint64(plt_entry[0:8]))  //This is the virtual address at which the linker will save the address for the searched function. This address is in the GOT, more specifically in the plt part of the GOT
			}
		}
	} else {
		rel_plt := get_section_by_name(data, curr_elf, ".rel.plt")
		rel_plt_section := data[rel_plt[0]:rel_plt[0]+rel_plt[1]]
		for index:=0;index < rel_plt[1]/rel_plt[2];index++ {  
			plt_entry := rel_plt_section[index*rel_plt[2]:(index+1)*rel_plt[2]]  
			info := int(binary.LittleEndian.Uint32(plt_entry[4:8])) >> 8  
			if info == id {
				return int(binary.LittleEndian.Uint32(plt_entry[0:4]))
			}
		}
	}

	return 0
}

func get_fun_addr_by_name(data []byte, curr_elf ELF, searched_function string) int{
	offset := strings.Index(curr_elf.string_table, searched_function)
	symtab := get_section_by_name(data, curr_elf, ".symtab")
	symtab_section := data[symtab[0]:symtab[0]+symtab[1]]
	for index:=0;index < symtab[1]/symtab[2];index++ {
		symtab_entry := symtab_section[index*symtab[2]:(index+1)*symtab[2]]
		if curr_elf.bit64 {
			name := int(binary.LittleEndian.Uint32(symtab_entry[0:4]))
			if name == offset {
				return int(binary.LittleEndian.Uint64(symtab_entry[8:16]))
			}
		} else {
			name := int(binary.LittleEndian.Uint32(symtab_entry[0:4]))
			if name == offset {
				return int(binary.LittleEndian.Uint32(symtab_entry[4:8]))
			}
		}
	
	}
	return 0

}

func vir_addr_to_phys_addr(data []byte, curr_elf ELF, vir_addr int, section_name string) int{  //translates virtual address to physical address based on section in which this address is
	section_info := get_section_by_name(data, curr_elf,section_name)  //get section information
	offset := vir_addr - section_info[3]  //subtracts the start address of the virtual section from the virtual address to get the offset into the section
	phys_addr := section_info[0] + offset  //adds this offset to the physical start of the section to get physical offset
	return phys_addr
}

func check_for_pie(data []byte) bool{  //checks if ELF file is PIE (position independent executable). If yes, there's no point in replacing entries in the GOT with predefined values
	if data[16] == 2 {  //This is the type field of the ELF header that specifies the type of an ELF file. 2 means EXEC which means that the ELF is not PIE. If it is, the value is 3 which stands for DYN
		return false
	} else {
		return true
	}
}

func check_for_64bit(data []byte) bool{  //parses ELF header to determine wether ELF is 64 bit or not
	if data[4] == 2 {
		return true
	} else {
		return false
	}
}

func overwrite_got_entry(data []byte, function_name string, new_function_address string, curr_elf ELF, is_hex bool) ([]byte) {  //overwrites GOT entry of a library function with a user specified function
	addr := get_dyn_addr_by_name(data, curr_elf, function_name)
	addr = vir_addr_to_phys_addr(data, curr_elf, addr, ".got.plt")
	if is_hex {
		data = write_data(new_function_address,data,addr,4)
	} else {
		new_function_address = fmt.Sprintf("%x",get_fun_addr_by_name(data, curr_elf,new_function_address))
		data = write_data(new_function_address,data,addr,4)
	}
	return data
}

func overwrite_section_header_types(data []byte, curr_elf ELF) ([]byte) {  //overwrites section types with null bytes
	for index := 0; index < curr_elf.number_of_sections; index++ {
		addr := curr_elf.section_table + index*curr_elf.size_of_section_headers + 4
		data = write_data("",data,addr,4)
	}
	return data
}

func overwrite_section_header_names(data []byte, curr_elf ELF) ([]byte) {  //overwrites section names with null bytes
	for index := 0; index < curr_elf.number_of_sections; index++ {
		addr := curr_elf.section_table + index*curr_elf.size_of_section_headers
		data = write_data("",data,addr,4)
	}
	return data
}

func check_stripped(data []byte, curr_elf ELF) (bool) {  //checks if file is stripped by searching for the .symtab section that isn't present in a stripped ELF file
	if get_section_by_name(data,curr_elf,".symtab")[0] == 0 {
		return true
	} else {
		return false
	}
}

func check_ELF(data []byte) bool {  //checks file header and compares it against and ELF header
	if data[0] == 0x7f && data[1] == 0x45 && data[2] == 0x4c && data[3] == 0x46 {
		return true
	} else {
		return false
	}
	
}


//func main() {
//	data := read_file("nopie32")
//	curr_elf := initialize_ELF(data)
//	fmt.Println(curr_elf)
//	addr := get_dyn_addr_by_name(data,curr_elf,"system")
//	fmt.Println(addr)
//	fmt.Println(get_section_by_name(data,curr_elf,".symtab"))
//}
