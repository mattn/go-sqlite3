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
  @file crypt.c
  Build strings, Tom St Denis
*/
#define NAME_VALUE(s) #s"="NAME(s)
#define NAME(s) #s

const char *crypt_build_settings =
   "LibTomCrypt " SCRYPT " (www.libtom.net)\n"
   "LibTomCrypt is public domain software.\n"
#if defined(INCLUDE_BUILD_DATE)
   "Built on " __DATE__ " at " __TIME__ "\n"
#endif
   "\n\nEndianness: "
#if defined(ENDIAN_NEUTRAL)
   "neutral/"
#endif
#if defined(ENDIAN_LITTLE)
   "little"
#elif defined(ENDIAN_BIG)
   "big"
#endif
   #if defined(ENDIAN_32BITWORD)
   " (32-bit words)\n"
   #elif defined(ENDIAN_64BITWORD)
   " (64-bit words)\n"
   #else
   " (no wordsize defined)\n"
   #endif
   "Clean stack: "
#if defined(LTC_CLEAN_STACK)
   "enabled\n"
#else
   "disabled\n"
#endif
   "\nCiphers built-in:\n"
#if defined(LTC_BLOWFISH)
   "   Blowfish\n"
#endif
#if defined(LTC_RC2)
   "   RC2\n"
#endif
#if defined(LTC_RC5)
   "   RC5\n"
#endif
#if defined(LTC_RC6)
   "   RC6\n"
#endif
#if defined(LTC_SAFERP)
   "   Safer+\n"
#endif
#if defined(LTC_SAFER)
   "   Safer\n"
#endif
#if defined(LTC_RIJNDAEL)
   "   Rijndael\n"
#endif
#if defined(LTC_XTEA)
   "   XTEA\n"
#endif
#if defined(LTC_TWOFISH)
   "   Twofish "
   #if defined(LTC_TWOFISH_SMALL) && defined(LTC_TWOFISH_TABLES) && defined(LTC_TWOFISH_ALL_TABLES)
       "(small, tables, all_tables)\n"
   #elif defined(LTC_TWOFISH_SMALL) && defined(LTC_TWOFISH_TABLES)
       "(small, tables)\n"
   #elif defined(LTC_TWOFISH_SMALL) && defined(LTC_TWOFISH_ALL_TABLES)
       "(small, all_tables)\n"
   #elif defined(LTC_TWOFISH_TABLES) && defined(LTC_TWOFISH_ALL_TABLES)
       "(tables, all_tables)\n"
   #elif defined(LTC_TWOFISH_SMALL)
       "(small)\n"
   #elif defined(LTC_TWOFISH_TABLES)
       "(tables)\n"
   #elif defined(LTC_TWOFISH_ALL_TABLES)
       "(all_tables)\n"
   #else
       "\n"
   #endif
#endif
#if defined(LTC_DES)
   "   DES\n"
#endif
#if defined(LTC_CAST5)
   "   CAST5\n"
#endif
#if defined(LTC_NOEKEON)
   "   Noekeon\n"
#endif
#if defined(LTC_SKIPJACK)
   "   Skipjack\n"
#endif
#if defined(LTC_KHAZAD)
   "   Khazad\n"
#endif
#if defined(LTC_ANUBIS)
   "   Anubis "
#endif
#if defined(LTC_ANUBIS_TWEAK)
   " (tweaked)"
#endif
   "\n"
#if defined(LTC_KSEED)
   "   KSEED\n"
#endif
#if defined(LTC_KASUMI)
   "   KASUMI\n"
#endif
#if defined(LTC_MULTI2)
   "   MULTI2\n"
#endif
#if defined(LTC_CAMELLIA)
   "   Camellia\n"
#endif
#if defined(LTC_IDEA)
   "   IDEA\n"
#endif
#if defined(LTC_SERPENT)
   "   Serpent\n"
#endif
   "Stream ciphers built-in:\n"
#if defined(LTC_CHACHA)
   "   ChaCha\n"
#endif
#if defined(LTC_SALSA20)
   "   Salsa20\n"
#endif
#if defined(LTC_XSALSA20)
   "   XSalsa20\n"
#endif
#if defined(LTC_SOSEMANUK)
   "   Sosemanuk\n"
#endif
#if defined(LTC_RABBIT)
   "   Rabbit\n"
#endif
#if defined(LTC_RC4_STREAM)
   "   RC4\n"
#endif
#if defined(LTC_SOBER128_STREAM)
   "   SOBER128\n"
#endif

    "\nHashes built-in:\n"
#if defined(LTC_SHA3)
   "   SHA3\n"
#endif
#if defined(LTC_KECCAK)
   "   KECCAK\n"
#endif
#if defined(LTC_SHA512)
   "   SHA-512\n"
#endif
#if defined(LTC_SHA384)
   "   SHA-384\n"
#endif
#if defined(LTC_SHA512_256)
   "   SHA-512/256\n"
#endif
#if defined(LTC_SHA256)
   "   SHA-256\n"
#endif
#if defined(LTC_SHA512_224)
   "   SHA-512/224\n"
#endif
#if defined(LTC_SHA224)
   "   SHA-224\n"
#endif
#if defined(LTC_TIGER)
   "   TIGER\n"
#endif
#if defined(LTC_SHA1)
   "   SHA1\n"
#endif
#if defined(LTC_MD5)
   "   MD5\n"
#endif
#if defined(LTC_MD4)
   "   MD4\n"
#endif
#if defined(LTC_MD2)
   "   MD2\n"
#endif
#if defined(LTC_RIPEMD128)
   "   RIPEMD128\n"
#endif
#if defined(LTC_RIPEMD160)
   "   RIPEMD160\n"
#endif
#if defined(LTC_RIPEMD256)
   "   RIPEMD256\n"
#endif
#if defined(LTC_RIPEMD320)
   "   RIPEMD320\n"
#endif
#if defined(LTC_WHIRLPOOL)
   "   WHIRLPOOL\n"
#endif
#if defined(LTC_BLAKE2S)
   "   BLAKE2S\n"
#endif
#if defined(LTC_BLAKE2B)
   "   BLAKE2B\n"
#endif
#if defined(LTC_CHC_HASH)
   "   CHC_HASH\n"
#endif

    "\nBlock Chaining Modes:\n"
#if defined(LTC_CFB_MODE)
    "   CFB\n"
#endif
#if defined(LTC_OFB_MODE)
    "   OFB\n"
#endif
#if defined(LTC_ECB_MODE)
    "   ECB\n"
#endif
#if defined(LTC_CBC_MODE)
    "   CBC\n"
#endif
#if defined(LTC_CTR_MODE)
    "   CTR\n"
#endif
#if defined(LTC_LRW_MODE)
    "   LRW"
#if defined(LTC_LRW_TABLES)
    " (tables) "
#endif
    "\n"
#endif
#if defined(LTC_F8_MODE)
    "   F8\n"
#endif
#if defined(LTC_XTS_MODE)
    "   XTS\n"
#endif

    "\nMACs:\n"
#if defined(LTC_HMAC)
    "   HMAC\n"
#endif
#if defined(LTC_OMAC)
    "   OMAC\n"
#endif
#if defined(LTC_PMAC)
    "   PMAC\n"
#endif
#if defined(LTC_PELICAN)
    "   PELICAN\n"
#endif
#if defined(LTC_XCBC)
    "   XCBC\n"
#endif
#if defined(LTC_F9_MODE)
    "   F9\n"
#endif
#if defined(LTC_POLY1305)
    "   POLY1305\n"
#endif
#if defined(LTC_BLAKE2SMAC)
    "   BLAKE2S MAC\n"
#endif
#if defined(LTC_BLAKE2BMAC)
    "   BLAKE2B MAC\n"
#endif

    "\nENC + AUTH modes:\n"
#if defined(LTC_EAX_MODE)
    "   EAX\n"
#endif
#if defined(LTC_OCB_MODE)
    "   OCB\n"
#endif
#if defined(LTC_OCB3_MODE)
    "   OCB3\n"
#endif
#if defined(LTC_CCM_MODE)
    "   CCM\n"
#endif
#if defined(LTC_GCM_MODE)
    "   GCM"
#if defined(LTC_GCM_TABLES)
    " (tables) "
#endif
#if defined(LTC_GCM_TABLES_SSE2)
    " (SSE2) "
#endif
   "\n"
#endif
#if defined(LTC_CHACHA20POLY1305_MODE)
    "   CHACHA20POLY1305\n"
#endif

    "\nPRNG:\n"
#if defined(LTC_YARROW)
    "   Yarrow ("NAME_VALUE(LTC_YARROW_AES)")\n"
#endif
#if defined(LTC_SPRNG)
    "   SPRNG\n"
#endif
#if defined(LTC_RC4)
    "   RC4\n"
#endif
#if defined(LTC_CHACHA20_PRNG)
    "   ChaCha20\n"
#endif
#if defined(LTC_FORTUNA)
    "   Fortuna (" NAME_VALUE(LTC_FORTUNA_POOLS) ", "
#if defined(LTC_FORTUNA_RESEED_RATELIMIT_TIMED)
    "LTC_FORTUNA_RESEED_RATELIMIT_TIMED, "
#else
    "LTC_FORTUNA_RESEED_RATELIMIT_STATIC, " NAME_VALUE(LTC_FORTUNA_WD)
#endif
    ")\n"
#endif
#if defined(LTC_SOBER128)
    "   SOBER128\n"
#endif

    "\nPK Crypto:\n"
#if defined(LTC_MRSA)
    "   RSA"
#if defined(LTC_RSA_BLINDING) && defined(LTC_RSA_CRT_HARDENING)
    " (with blinding and CRT hardening)"
#elif defined(LTC_RSA_BLINDING)
    " (with blinding)"
#elif defined(LTC_RSA_CRT_HARDENING)
    " (with CRT hardening)"
#endif
    "\n"
#endif
#if defined(LTC_MDH)
    "   DH\n"
#endif
#if defined(LTC_MECC)
    "   ECC"
#if defined(LTC_ECC_TIMING_RESISTANT)
    " (with blinding)"
#endif
    "\n"
#endif
#if defined(LTC_MDSA)
    "   DSA\n"
#endif
#if defined(LTC_PK_MAX_RETRIES)
    "   "NAME_VALUE(LTC_PK_MAX_RETRIES)"\n"
#endif

    "\nMPI (Math):\n"
#if defined(LTC_MPI)
    "   LTC_MPI\n"
#endif
#if defined(LTM_DESC)
    "   LTM_DESC\n"
#endif
#if defined(TFM_DESC)
    "   TFM_DESC\n"
#endif
#if defined(GMP_DESC)
    "   GMP_DESC\n"
#endif
#if defined(LTC_MILLER_RABIN_REPS)
    "   "NAME_VALUE(LTC_MILLER_RABIN_REPS)"\n"
#endif

    "\nCompiler:\n"
#if defined(_WIN64)
    "   WIN64 platform detected.\n"
#elif defined(_WIN32)
    "   WIN32 platform detected.\n"
#endif
#if defined(__CYGWIN__)
    "   CYGWIN Detected.\n"
#endif
#if defined(__DJGPP__)
    "   DJGPP Detected.\n"
#endif
#if defined(_MSC_VER)
    "   MSVC compiler detected.\n"
#endif
#if defined(__clang_version__)
    "   Clang compiler " __clang_version__ ".\n"
#elif defined(INTEL_CC)
    "   Intel C Compiler " __VERSION__ ".\n"
#elif defined(__GNUC__)         /* clang and icc also define __GNUC__ */
    "   GCC compiler " __VERSION__ ".\n"
#endif

#if defined(__x86_64__)
    "   x86-64 detected.\n"
#endif
#if defined(LTC_PPC32)
    "   PPC32 detected.\n"
#endif

    "\nVarious others: "
#if defined(ARGTYPE)
    " " NAME_VALUE(ARGTYPE) " "
#endif
#if defined(LTC_ADLER32)
    " ADLER32 "
#endif
#if defined(LTC_BASE64)
    " BASE64 "
#endif
#if defined(LTC_BASE64_URL)
    " BASE64-URL-SAFE "
#endif
#if defined(LTC_BASE32)
    " BASE32 "
#endif
#if defined(LTC_BASE16)
    " BASE16 "
#endif
#if defined(LTC_CRC32)
    " CRC32 "
#endif
#if defined(LTC_DER)
    " DER "
    " " NAME_VALUE(LTC_DER_MAX_RECURSION) " "
#endif
#if defined(LTC_PKCS_1)
    " PKCS#1 "
#endif
#if defined(LTC_PKCS_5)
    " PKCS#5 "
#endif
#if defined(LTC_PKCS_12)
    " PKCS#12 "
#endif
#if defined(LTC_PADDING)
    " PADDING "
#endif
#if defined(LTC_HKDF)
    " HKDF "
#endif
#if defined(LTC_DEVRANDOM)
    " LTC_DEVRANDOM "
#endif
#if defined(LTC_TRY_URANDOM_FIRST)
    " LTC_TRY_URANDOM_FIRST "
#endif
#if defined(LTC_RNG_GET_BYTES)
    " LTC_RNG_GET_BYTES "
#endif
#if defined(LTC_RNG_MAKE_PRNG)
    " LTC_RNG_MAKE_PRNG "
#endif
#if defined(LTC_PRNG_ENABLE_LTC_RNG)
    " LTC_PRNG_ENABLE_LTC_RNG "
#endif
#if defined(LTC_HASH_HELPERS)
    " LTC_HASH_HELPERS "
#endif
#if defined(LTC_VALGRIND)
    " LTC_VALGRIND "
#endif
#if defined(LTC_TEST)
    " LTC_TEST "
#endif
#if defined(LTC_TEST_DBG)
    " " NAME_VALUE(LTC_TEST_DBG) " "
#endif
#if defined(LTC_TEST_EXT)
    " LTC_TEST_EXT "
#endif
#if defined(LTC_SMALL_CODE)
    " LTC_SMALL_CODE "
#endif
#if defined(LTC_NO_FILE)
    " LTC_NO_FILE "
#endif
#if defined(LTC_FILE_READ_BUFSIZE)
    " " NAME_VALUE(LTC_FILE_READ_BUFSIZE) " "
#endif
#if defined(LTC_FAST)
    " LTC_FAST "
#endif
#if defined(LTC_NO_FAST)
    " LTC_NO_FAST "
#endif
#if defined(LTC_NO_BSWAP)
    " LTC_NO_BSWAP "
#endif
#if defined(LTC_NO_ASM)
    " LTC_NO_ASM "
#endif
#if defined(LTC_ROx_ASM)
    " LTC_ROx_ASM "
#if defined(LTC_NO_ROLC)
    " LTC_NO_ROLC "
#endif
#endif
#if defined(LTC_NO_TEST)
    " LTC_NO_TEST "
#endif
#if defined(LTC_NO_TABLES)
    " LTC_NO_TABLES "
#endif
#if defined(LTC_PTHREAD)
    " LTC_PTHREAD "
#endif
#if defined(LTC_EASY)
    " LTC_EASY "
#endif
#if defined(LTC_MECC_ACCEL)
    " LTC_MECC_ACCEL "
#endif
#if defined(LTC_MECC_FP)
    " LTC_MECC_FP "
#endif
#if defined(LTC_ECC_SHAMIR)
    " LTC_ECC_SHAMIR "
#endif
#if defined(LTC_CLOCK_GETTIME)
    " LTC_CLOCK_GETTIME "
#endif
    "\n"
    ;


/* ref:         $Format:%D$ */
/* git commit:  $Format:%H$ */
/* commit time: $Format:%ai$ */
