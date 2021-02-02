gcc example.c -o example #normal compilation
gcc -s example.c -o example_stripped #stripped compilation
gcc -no-pie example.c -o example_nopie #no pie compilation
gcc -m32 example.c -o example_32 #32 bit compilation, may need external libraries
