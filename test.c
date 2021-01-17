#include <stdio.h>
#include <stdlib.h>


int custom_printf(char* input)
{
	printf("This is custom\n");
	printf(input);
	return 1;
}

int main()
{
	int x = system("Test\n");
	printf("%i\n",x);
}
