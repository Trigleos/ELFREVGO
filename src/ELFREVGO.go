package main

import (
	"flag"
	"fmt"
	"strings"
)

/// -e change class -b change bits -t overwrite section types -n overwrite section names -g overwrite got function -gd destination function -gf new function -gx new address -f filename -o output filename -i interactive


func main() {
	
	change_endian_ptr := flag.Bool("e",false,"change endianness of ELF")
	change_bits_ptr := flag.Bool("b",false,"change number of bits (32 or 64) of ELF")
	overwrite_sec_types := flag.Bool("t", false, "overwrite section types with null bytes")
	overwrite_sec_names := flag.Bool("n", false, "overwrite section names with null bytes")
	overwrite_got_function := flag.Bool("g", false, "overwrite library function with another function")
	dest_func := flag.String("gd","","name of the library functions that you want to replace, separated by commas")
	new_func := flag.String("gf","","name of the functions that you want to call instead of the library functions in the same order as the library functions, separated by commas")
	new_hex := flag.String("gx","","hexadecimal addresses that you want to call instead of the library functions in the same order as the library functions, separated by commas")
	filename := flag.String("f","","name of the ELF file you want to change")
	output_filename := flag.String("o","","name of output ELF file")
	
	flag.Parse()
	
	if len(*filename) == 0 {
		fmt.Println("Please provide a filename")
		return
	}
	
	if len(*output_filename) == 0 {
		*output_filename = *filename + "_go"
	}
	
	
	data := read_file(*filename)
	if len(data) == 0 {
		return
	}
	
	if !check_ELF(data) {
		fmt.Println("This file doesn't seem to be an ELF file")
		return
	}
	
	curr_elf := initialize_ELF(data)
	
	if(*overwrite_got_function) {
		if len(*dest_func) == 0 {
			fmt.Println("Please provide a library function name to replace")
			return
		} else if len(*new_func) == 0 && len(*new_hex) == 0 {
			fmt.Println("Please provide a function to replace the library function with")
			return
		}
		
		if check_for_pie(data) {
			fmt.Println("ELF file seems to be a PIE, replacing a GOT entry won't achieve the desired effect. Compile the script with the -no-pie flag on gcc to disable PIE")
			return
		} else {
			if len(*new_func) != 0 && check_stripped(data, curr_elf) {
				fmt.Println("ELF file seems to be stripped thus it is not possible to determine the address of the specified function.\nPlease enter the function address in hex instead")
				return
			} else if len(*new_func) != 0 {
				lib_funcs := strings.Split(*dest_func,",")
				new_funcs := strings.Split(*new_func,",")
				for index := 0;index < len(lib_funcs); index++ {
					data = overwrite_got_entry(data,lib_funcs[index],new_funcs[index],curr_elf, false)
				}
				
			} else {
				lib_funcs := strings.Split(*dest_func,",")
				new_funcs := strings.Split(*new_hex,",")
				for index := 0;index < len(lib_funcs); index++ {
					data = overwrite_got_entry(data,lib_funcs[index],new_funcs[index],curr_elf, true)
				}
			}
		}
	}
	
	if(*overwrite_sec_types) {
		data = overwrite_section_header_types(data, curr_elf)
	}
	
	if(*overwrite_sec_names) {
		data = overwrite_section_header_names(data, curr_elf)
	}
	
	if(*change_endian_ptr) {
		data = change_endianness(data)
	}
	
	if(*change_bits_ptr) {
		data = change_class(data)
	}
	
	write_file(*output_filename,data)
	
	
	
	//data = overwrite_got_entry(data, "system" , "custom_printf",curr_elf, false)
	
	//data = overwrite_section_header_types(data, curr_elf)
	
	//write_file("nopie_go",data)

}
