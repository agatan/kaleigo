extern double __kaleigo_main();
#include <stdio.h>

double putd(double d) {
  printf("%f\n", d);
  return 0.0;
}

double putchard(double d) {
  putchar((int)d);
  return  0.0;
}

int main(void) {
  __kaleigo_main();
}
