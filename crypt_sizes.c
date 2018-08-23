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
  @file crypt_sizes.c

  Make various struct sizes available to dynamic languages
  like Python - Larry Bugbee, February 2013

  LB - Dec 2013 - revised to include compiler define options
*/


typedef struct {
    const char *name;
    const unsigned int size;
} crypt_size;

#define _SZ_STRINGIFY_S(s) { #s, sizeof(struct s) }
#define _SZ_STRINGIFY_T(s) { #s, sizeof(s) }

static const crypt_size _crypt_sizes[] = {
    /* hash state sizes */
    _SZ_STRINGIFY_S(ltc_hash_descriptor),
    _SZ_STRINGIFY_T(hash_state),
#ifdef LTC_CHC_HASH
    _SZ_STRINGIFY_S(chc_state),
#endif
#ifdef LTC_WHIRLPOOL
    _SZ_STRINGIFY_S(whirlpool_state),
#endif
#ifdef LTC_SHA3
    _SZ_STRINGIFY_S(sha3_state),
#endif
#ifdef LTC_SHA512
    _SZ_STRINGIFY_S(sha512_state),
#endif
#ifdef LTC_SHA256
    _SZ_STRINGIFY_S(sha256_state),
#endif
#ifdef LTC_SHA1
    _SZ_STRINGIFY_S(sha1_state),
#endif
#ifdef LTC_MD5
    _SZ_STRINGIFY_S(md5_state),
#endif
#ifdef LTC_MD4
    _SZ_STRINGIFY_S(md4_state),
#endif
#ifdef LTC_MD2
    _SZ_STRINGIFY_S(md2_state),
#endif
#ifdef LTC_TIGER
    _SZ_STRINGIFY_S(tiger_state),
#endif
#ifdef LTC_RIPEMD128
    _SZ_STRINGIFY_S(rmd128_state),
#endif
#ifdef LTC_RIPEMD160
    _SZ_STRINGIFY_S(rmd160_state),
#endif
#ifdef LTC_RIPEMD256
    _SZ_STRINGIFY_S(rmd256_state),
#endif
#ifdef LTC_RIPEMD320
    _SZ_STRINGIFY_S(rmd320_state),
#endif
#ifdef LTC_BLAKE2S
    _SZ_STRINGIFY_S(blake2s_state),
#endif
#ifdef LTC_BLAKE2B
    _SZ_STRINGIFY_S(blake2b_state),
#endif

    /* block cipher key sizes */
    _SZ_STRINGIFY_S(ltc_cipher_descriptor),
    _SZ_STRINGIFY_T(symmetric_key),
#ifdef LTC_ANUBIS
    _SZ_STRINGIFY_S(anubis_key),
#endif
#ifdef LTC_CAMELLIA
    _SZ_STRINGIFY_S(camellia_key),
#endif
#ifdef LTC_BLOWFISH
    _SZ_STRINGIFY_S(blowfish_key),
#endif
#ifdef LTC_CAST5
    _SZ_STRINGIFY_S(cast5_key),
#endif
#ifdef LTC_DES
    _SZ_STRINGIFY_S(des_key),
    _SZ_STRINGIFY_S(des3_key),
#endif
#ifdef LTC_IDEA
    _SZ_STRINGIFY_S(idea_key),
#endif
#ifdef LTC_KASUMI
    _SZ_STRINGIFY_S(kasumi_key),
#endif
#ifdef LTC_KHAZAD
    _SZ_STRINGIFY_S(khazad_key),
#endif
#ifdef LTC_KSEED
    _SZ_STRINGIFY_S(kseed_key),
#endif
#ifdef LTC_MULTI2
    _SZ_STRINGIFY_S(multi2_key),
#endif
#ifdef LTC_NOEKEON
    _SZ_STRINGIFY_S(noekeon_key),
#endif
#ifdef LTC_RC2
    _SZ_STRINGIFY_S(rc2_key),
#endif
#ifdef LTC_RC5
    _SZ_STRINGIFY_S(rc5_key),
#endif
#ifdef LTC_RC6
    _SZ_STRINGIFY_S(rc6_key),
#endif
#ifdef LTC_SERPENT
    _SZ_STRINGIFY_S(serpent_key),
#endif
#ifdef LTC_SKIPJACK
    _SZ_STRINGIFY_S(skipjack_key),
#endif
#ifdef LTC_XTEA
    _SZ_STRINGIFY_S(xtea_key),
#endif
#ifdef LTC_RIJNDAEL
    _SZ_STRINGIFY_S(rijndael_key),
#endif
#ifdef LTC_SAFER
    _SZ_STRINGIFY_S(safer_key),
#endif
#ifdef LTC_SAFERP
    _SZ_STRINGIFY_S(saferp_key),
#endif
#ifdef LTC_TWOFISH
    _SZ_STRINGIFY_S(twofish_key),
#endif

    /* mode sizes */
#ifdef LTC_ECB_MODE
    _SZ_STRINGIFY_T(symmetric_ECB),
#endif
#ifdef LTC_CFB_MODE
    _SZ_STRINGIFY_T(symmetric_CFB),
#endif
#ifdef LTC_OFB_MODE
    _SZ_STRINGIFY_T(symmetric_OFB),
#endif
#ifdef LTC_CBC_MODE
    _SZ_STRINGIFY_T(symmetric_CBC),
#endif
#ifdef LTC_CTR_MODE
    _SZ_STRINGIFY_T(symmetric_CTR),
#endif
#ifdef LTC_LRW_MODE
    _SZ_STRINGIFY_T(symmetric_LRW),
#endif
#ifdef LTC_F8_MODE
    _SZ_STRINGIFY_T(symmetric_F8),
#endif
#ifdef LTC_XTS_MODE
    _SZ_STRINGIFY_T(symmetric_xts),
#endif

    /* stream cipher sizes */
#ifdef LTC_CHACHA
    _SZ_STRINGIFY_T(chacha_state),
#endif
#ifdef LTC_SALSA20
    _SZ_STRINGIFY_T(salsa20_state),
#endif
#ifdef LTC_SOSEMANUK
    _SZ_STRINGIFY_T(sosemanuk_state),
#endif
#ifdef LTC_RABBIT
    _SZ_STRINGIFY_T(rabbit_state),
#endif
#ifdef LTC_RC4_STREAM
    _SZ_STRINGIFY_T(rc4_state),
#endif
#ifdef LTC_SOBER128_STREAM
    _SZ_STRINGIFY_T(sober128_state),
#endif

    /* MAC sizes            -- no states for ccm, lrw */
#ifdef LTC_HMAC
    _SZ_STRINGIFY_T(hmac_state),
#endif
#ifdef LTC_OMAC
    _SZ_STRINGIFY_T(omac_state),
#endif
#ifdef LTC_PMAC
    _SZ_STRINGIFY_T(pmac_state),
#endif
#ifdef LTC_POLY1305
    _SZ_STRINGIFY_T(poly1305_state),
#endif
#ifdef LTC_EAX_MODE
    _SZ_STRINGIFY_T(eax_state),
#endif
#ifdef LTC_OCB_MODE
    _SZ_STRINGIFY_T(ocb_state),
#endif
#ifdef LTC_OCB3_MODE
    _SZ_STRINGIFY_T(ocb3_state),
#endif
#ifdef LTC_CCM_MODE
    _SZ_STRINGIFY_T(ccm_state),
#endif
#ifdef LTC_GCM_MODE
    _SZ_STRINGIFY_T(gcm_state),
#endif
#ifdef LTC_PELICAN
    _SZ_STRINGIFY_T(pelican_state),
#endif
#ifdef LTC_XCBC
    _SZ_STRINGIFY_T(xcbc_state),
#endif
#ifdef LTC_F9_MODE
    _SZ_STRINGIFY_T(f9_state),
#endif
#ifdef LTC_CHACHA20POLY1305_MODE
    _SZ_STRINGIFY_T(chacha20poly1305_state),
#endif

    /* asymmetric keys */
#ifdef LTC_MRSA
    _SZ_STRINGIFY_T(rsa_key),
#endif
#ifdef LTC_MDSA
    _SZ_STRINGIFY_T(dsa_key),
#endif
#ifdef LTC_MDH
    _SZ_STRINGIFY_T(dh_key),
#endif
#ifdef LTC_MECC
    _SZ_STRINGIFY_T(ltc_ecc_curve),
    _SZ_STRINGIFY_T(ecc_point),
    _SZ_STRINGIFY_T(ecc_key),
#endif

    /* DER handling */
#ifdef LTC_DER
    _SZ_STRINGIFY_T(ltc_asn1_list),  /* a list entry */
    _SZ_STRINGIFY_T(ltc_utctime),
    _SZ_STRINGIFY_T(ltc_generalizedtime),
#endif

    /* prng state sizes */
    _SZ_STRINGIFY_S(ltc_prng_descriptor),
    _SZ_STRINGIFY_T(prng_state),
#ifdef LTC_FORTUNA
    _SZ_STRINGIFY_S(fortuna_prng),
#endif
#ifdef LTC_CHACHA20_PRNG
    _SZ_STRINGIFY_S(chacha20_prng),
#endif
#ifdef LTC_RC4
    _SZ_STRINGIFY_S(rc4_prng),
#endif
#ifdef LTC_SOBER128
    _SZ_STRINGIFY_S(sober128_prng),
#endif
#ifdef LTC_YARROW
    _SZ_STRINGIFY_S(yarrow_prng),
#endif
    /* sprng has no state as it uses other potentially available sources */
    /* like /dev/random.  See Developers Guide for more info. */

#ifdef LTC_ADLER32
    _SZ_STRINGIFY_T(adler32_state),
#endif
#ifdef LTC_CRC32
    _SZ_STRINGIFY_T(crc32_state),
#endif

    _SZ_STRINGIFY_T(ltc_mp_digit),
    _SZ_STRINGIFY_T(ltc_math_descriptor)

};

/* crypt_get_size()
 * sizeout will be the size (bytes) of the named struct or union
 * return -1 if named item not found
 */
int crypt_get_size(const char* namein, unsigned int *sizeout) {
    int i;
    int count = sizeof(_crypt_sizes) / sizeof(_crypt_sizes[0]);
    for (i=0; i<count; i++) {
        if (XSTRCMP(_crypt_sizes[i].name, namein) == 0) {
            *sizeout = _crypt_sizes[i].size;
            return 0;
        }
    }
    return -1;
}

/* crypt_list_all_sizes()
 * if names_list is NULL, names_list_size will be the minimum
 *     size needed to receive the complete names_list
 * if names_list is NOT NULL, names_list must be the addr with
 *     sufficient memory allocated into which the names_list
 *     is to be written.  Also, the value in names_list_size
 *     sets the upper bound of the number of characters to be
 *     written.
 * a -1 return value signifies insufficient space made available
 */
int crypt_list_all_sizes(char *names_list, unsigned int *names_list_size) {
    int i;
    unsigned int total_len = 0;
    char *ptr;
    int number_len;
    int count = sizeof(_crypt_sizes) / sizeof(_crypt_sizes[0]);

    /* calculate amount of memory required for the list */
    for (i=0; i<count; i++) {
        number_len = snprintf(NULL, 0, "%s,%u\n", _crypt_sizes[i].name, _crypt_sizes[i].size);
        if (number_len < 0) {
          return -1;
        }
        total_len += number_len;
        /* this last +1 is for newlines (and ending NULL) */
    }

    if (names_list == NULL) {
        *names_list_size = total_len;
    } else {
        if (total_len > *names_list_size) {
            return -1;
        }
        /* build the names list */
        ptr = names_list;
        for (i=0; i<count; i++) {
            number_len = snprintf(ptr, total_len, "%s,%u\n", _crypt_sizes[i].name, _crypt_sizes[i].size);
            if (number_len < 0) return -1;
            if ((unsigned int)number_len > total_len) return -1;
            total_len -= number_len;
            ptr += number_len;
        }
        /* to remove the trailing new-line */
        ptr -= 1;
        *ptr = 0;
    }
    return 0;
}


/* ref:         $Format:%D$ */
/* git commit:  $Format:%H$ */
/* commit time: $Format:%ai$ */
