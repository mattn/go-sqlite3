/* LibTomCrypt, modular cryptographic library -- Tom St Denis
 *
 * LibTomCrypt is a library that provides various cryptographic
 * algorithms in a highly modular and flexible manner.
 *
 * The library is free for all purposes without any express
 * guarantee it works.
 */
#include "tomcrypt_private.h"

/**
  @file crypt_inits.c

  Provide math library functions for dynamic languages
  like Python - Larry Bugbee, February 2013
*/


#ifdef LTM_DESC
void init_LTM(void)
{
    ltc_mp = ltm_desc;
}
#endif

#ifdef TFM_DESC
void init_TFM(void)
{
    ltc_mp = tfm_desc;
}
#endif

#ifdef GMP_DESC
void init_GMP(void)
{
    ltc_mp = gmp_desc;
}
#endif

int crypt_mp_init(const char* mpi)
{
   if (mpi == NULL) return CRYPT_ERROR;
   switch (mpi[0]) {
#ifdef LTM_DESC
      case 'l':
      case 'L':
         ltc_mp = ltm_desc;
         return CRYPT_OK;
#endif
#ifdef TFM_DESC
      case 't':
      case 'T':
         ltc_mp = tfm_desc;
         return CRYPT_OK;
#endif
#ifdef GMP_DESC
      case 'g':
      case 'G':
         ltc_mp = gmp_desc;
         return CRYPT_OK;
#endif
#ifdef EXT_MATH_LIB
      case 'e':
      case 'E':
         {
            extern ltc_math_descriptor EXT_MATH_LIB;
            ltc_mp = EXT_MATH_LIB;
         }

#if defined(LTC_TEST_DBG)
#define NAME_VALUE(s) #s"="NAME(s)
#define NAME(s) #s
         printf("EXT_MATH_LIB = %s\n", NAME_VALUE(EXT_MATH_LIB));
#undef NAME_VALUE
#undef NAME
#endif

         return CRYPT_OK;
#endif
      default:
#if defined(LTC_TEST_DBG)
         printf("Unknown/Invalid MPI provider: %s\n", mpi);
#endif
         return CRYPT_ERROR;
   }
}


/* ref:         $Format:%D$ */
/* git commit:  $Format:%H$ */
/* commit time: $Format:%ai$ */
