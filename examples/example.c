#include <stdio.h>
#include <stdlib.h>


int custom_printf(char* input) /*This is the function we want to call instead of system*/
{
	printf("This is custom\n");
	printf("%s",input);
	return 1;
}

int main()
{
	int x = system("Test\n");  /*This is the call to system we want to replace*/
	printf("%i\n",x);
}
