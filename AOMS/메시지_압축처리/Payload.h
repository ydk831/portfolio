#ifndef __Payload_H__
#define __Payload_H__ 
#include <bits/types.h>
#include <string>
#include <unistd.h>
#include <stdint.h>
#include "zlib.h"

using namespace std;

class Payload {
public:
	char               *GetDataStr() { return m_contentData; }
	char               *GetDataStrConst() const { return m_contentData; }
	__uint32_t          GetDataSize() { return m_contentSize; }
	__uint32_t          GetDataSizeConst() const { return m_contentSize; }
	void                SetDataSize(__uint32_t size) { m_contentSize = size; }

	void                Reset(__uint32_t size);
	Payload            *Clone();
	void                Clear();
	void operator=(const Payload& src);

	Payload();
	Payload(const char *pBuf, const __uint32_t size);
	~Payload();

private:
	char               *m_contentData;
	__uint32_t          m_contentSize;
};


//R440
class CmpPayload {
public:
	char               *GetDataStr() { return m_contentData; }
	char               *GetDataStrConst() const { return m_contentData; }

	__uint32_t          GetCmpSize() { return m_compressSize; }
	__uint32_t          GetCmpSizeConst() const { return m_compressSize; }

	__uint32_t          GetEncSize() { return m_encodingSize; }
	__uint32_t          GetEncSizeConst() const { return m_encodingSize; }

	void                SetCmpSize(__uint32_t cmp) { m_compressSize = cmp; }
	void                SetEncSize(__uint32_t enc) { m_encodingSize = enc; }

	void                Reset(__uint32_t size);
	CmpPayload         *Clone();
	void                Clear();
	void                operator=(const CmpPayload& src);
	void                operator=(const Payload& src);

	void                CompressPayload(char *buf, __uint32_t size);
	bool                GetOrgPayload(Payload &dst);

	int                 GetMaxCompressedLen(int nLenSrc);
	int                 CompressData(const Bytef* abSrc, int nLenSrc, Bytef* abDst, int nLenDst);
	int                 UncompressData(const Bytef* abSrc, int nLenSrc, Bytef** abDst, __uint32_t &rLenDst);

	bool                Encode_CmpPayload_Base64(unsigned char *cmp_data);
	bool                Decode_CmpPayload_Base64(char *&pBuf);

	CmpPayload();
	CmpPayload(const char *pBuf, const __uint32_t size);
	CmpPayload(const char *pBuf, const __uint32_t cmp, const __uint32_t enc);
	CmpPayload(Payload &payload);
	CmpPayload(CmpPayload &payload);
	~CmpPayload();

private:
	char               *m_contentData;
	__uint32_t          m_compressSize;
	__uint32_t          m_encodingSize;
};

#endif
