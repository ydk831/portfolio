#include <string.h>
#include "Payload.h"
#include <zutil.h>

#ifndef _SIM_
#include "CommIh/DebugLogger.h"
#include "base64.h"
#endif

void CmpPayload::CompressPayload(char *buf, __uint32_t size)
{
	static const char *FN = "[CmpPayload::CompressPayload] ";
	__uint32_t src_size = size;
	int cmp_max_size = GetMaxCompressedLen(src_size);

	byte* cmp_data = (byte*)malloc(cmp_max_size);
	int cmp_length = CompressData((byte*)buf, src_size, cmp_data, cmp_max_size);

	if (cmp_length == -1) {
		free(cmp_data);
		EELOG(RED(FN << " Compress Payload Fail."));
	}
	else {
		m_compressSize = cmp_length;
		DDLOG(CYAN(FN << "Payload Compression Success. Original Size : " << size << ", Compress Size : " << m_compressSize));

		if (!(Encode_CmpPayload_Base64(cmp_data))) {
			EELOG(RED(FN << "Payload Encoding Fail"));
		}

		free(cmp_data);
	}
}

bool CmpPayload::GetOrgPayload(Payload &dst)
{
	static const char *FN = "[CmpPayload::GetOrgPayload] ";

	if (!m_contentData && !m_compressSize) {
		EELOG(RED(FN << "CmpPayload Data Empty"));
		return false;
	}

	char *pBuf = NULL;
	if (Decode_CmpPayload_Base64(pBuf) == false) {
		EELOG(RED(FN << "CmpPayload Decode Fail"));
		return false;
	}
	else if (pBuf == NULL) {
		EELOG(RED(FN << "CmpPayload pBuf NULL"));
		return false;
	}

	byte* dcmp_data = NULL;
	__uint32_t dcmp_size = 0;
	int dcmp_length;

	byte* cmp_data = (byte*)malloc(5120);
	memcpy(cmp_data, pBuf, m_compressSize);

	dcmp_length = UncompressData(cmp_data, m_compressSize, (Bytef**)&dcmp_data, dcmp_size);
	if (dcmp_length == -1) {
		free(pBuf);
		free(cmp_data);
		EELOG(RED(FN << "Payload DeCompression Fail"));
		return false;
	}
	else {
		char buf[dcmp_length + 1];
		buf[dcmp_length] = 0;
		memcpy(buf, dcmp_data, dcmp_length);

		DDLOG(CYAN(FN << "Payload DeCompression Success. Compress Encoding Size : " << m_encodingSize << ", original Size : " << dcmp_length));
		free(pBuf);
		free(cmp_data);

		Payload payload(buf, dcmp_length);
		dst = payload;
		return true;
	}
}

int CmpPayload::CompressData(const Bytef* abSrc, int nLenSrc, Bytef* abDst, int nLenDst)
{
	z_stream zInfo;
	memset(&zInfo, 0x00, sizeof(zInfo));

	zInfo.total_in = 0;
	zInfo.total_out = 0;
	zInfo.avail_in = nLenSrc;
	zInfo.avail_out = nLenDst;
	zInfo.next_in = (Bytef*)abSrc;
	zInfo.next_out = abDst;

	int nErr, nRet = -1;
	nErr = deflateInit(&zInfo, Z_BEST_COMPRESSION);     /* 단말에서 사용하는 압축 모드 그대로. */
	if (nErr == Z_OK) {
		nErr = deflate(&zInfo, Z_FINISH);
		if (nErr == Z_STREAM_END) {
			nRet = zInfo.total_out;
		}
	}
	deflateEnd(&zInfo);

	return(nRet);
}

int CmpPayload::UncompressData(const Bytef* abSrc, int nLenSrc, Bytef** abDst, __uint32_t &rLenDst)
{
	static const char *FN = "CmpPayload::UncompressData] ";

	if (*abDst != NULL) {
		free(*abDst);
		*abDst = NULL;
		rLenDst = 0;
	}

	z_stream    d_stream;
	int         err, idx = 0, UNIT = 1024;
	char        buffer[UNIT];

	d_stream.zalloc = Z_NULL;
	d_stream.zfree = Z_NULL;
	d_stream.opaque = Z_NULL;

	d_stream.next_in = (Bytef*)abSrc;
	d_stream.avail_in = 0;

	err = inflateInit(&d_stream);
	if (err != Z_OK) {
		EELOG(RED(FN << "inflateInit fail [" << err << "]"));
		return -1;
	}

	*abDst = (unsigned char*)malloc(UNIT);
	rLenDst = UNIT;

	while (d_stream.total_in < (__u_int)nLenSrc) {
		d_stream.total_out = 0;
		d_stream.next_out = (Bytef*)buffer;
		d_stream.avail_in = d_stream.avail_out = UNIT;
		err = inflate(&d_stream, Z_NO_FLUSH);
		if (d_stream.total_out > 0) {
			memcpy(*abDst + idx, buffer, d_stream.total_out);
			idx += d_stream.total_out;
			*abDst = (unsigned char*)realloc(*abDst, rLenDst + UNIT);
			rLenDst += UNIT;
		}
		if (err == Z_STREAM_END) {
			break;
		}
		if (err != Z_OK) {
			EELOG(RED(FN << "inflate error [" << err << "]"));
			return -1;
		}
	}

	err = inflateEnd(&d_stream);
	if (err != Z_OK) {
		EELOG(RED(FN << "inflateEnd fail [" << err << "]"));
		return -1;
	}

	return idx;
}

bool CmpPayload::Encode_CmpPayload_Base64(unsigned char *cmp_data)
{
	const static char *FN = "[CmpPayload::Encode_CmpPayload_Base64] ";
	static const int32_t nChunkLen = 5120;
	bool bRet = true;
	int32_t nLen = m_compressSize;
	int32_t nBase64StrLen = (nLen + 2 - ((nLen + 2) % 3)) / 3 * 4;
	int32_t nBase64BufferLen = (nBase64StrLen + nChunkLen) & ~(nChunkLen - 1);
	unsigned char *pBase64Buffer = new unsigned char[nBase64BufferLen];
	if (pBase64Buffer == NULL)
	{
		EELOG(FN << "Alloc Fail");
		return false;
	}

	Base64 Encoder;
	string rsDst = "";

	if (0 > Encoder.encode(pBase64Buffer, (unsigned int*)&nBase64BufferLen, (unsigned char *)cmp_data, (unsigned int)m_compressSize))
	{
		bRet = false;
		goto cleanup;
	}

	rsDst += (char *)pBase64Buffer;
	m_encodingSize = rsDst.length();

	m_contentData = new char[m_encodingSize + 1];
	m_contentData[m_encodingSize] = 0;
	memcpy(m_contentData, rsDst.c_str(), rsDst.length());
	DDLOG(CYAN(FN << "CmpPayload Encoding Success. Compress Encoding Size: " << m_encodingSize));

cleanup:
	delete[] pBase64Buffer;
	return bRet;
}

bool CmpPayload::Decode_CmpPayload_Base64(char *&pBuf)
{
	const static char *FN = "CmpPayload::Decode_CmpPayload_Base64] ";

	if (m_contentData == NULL) {
		EELOG(RED(FN << " m_contentData is NULL"));
		return false;
	}

	static const int32_t nChunkLen = 5120;
	bool bRet = true;

	string sBase64Str(m_contentData);
	int32_t nLen = sBase64Str.size();
	int32_t nPayloadBufferLen = (nLen + nChunkLen) & ~(nChunkLen - 1);

	pBuf = new char[nPayloadBufferLen + 1];
	pBuf[nPayloadBufferLen] = 0;

	Base64 Decoder;
	if (0 > Decoder.decode((unsigned char *)pBuf, (unsigned int *)&nPayloadBufferLen, (unsigned char *)sBase64Str.c_str(), (unsigned int)sBase64Str.size()))
	{
		bRet = false;
		delete pBuf;
		pBuf = NULL;
		EELOG(RED(FN << " Compress Payload Base64 Decoding Fail!!"));
		return bRet;
	}

	return bRet;
}