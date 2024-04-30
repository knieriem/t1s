#include "lib/oa-tc6/src/tc6-regs.c"
#include "lib/oa-tc6/src/tc6.c"

#if 0
#	include "_cgo_export.h"
#	undef uint64_t
#	define castUint64 (GoUint64*)
#else
#	define castUint64

extern	void	tc6_onNeedService(TC6_t *pInst, void *pGlobalTag);

extern	void	tc6_onError(TC6_t *pInst, TC6_Error_t e, void *pGlobalTag);

extern	void	tc6_onRxEthernetSlice(TC6_t*, const uint8_t *pRx, uint16_t offset, uint16_t len, void *pGlobalTag);

extern	void	tc6_onRxEthernetPacket(TC6_t*, int success, uint16_t len, uint64_t *rxTimestamp, void *pGlobalTag);

extern	void	tc6regs_onEvent(TC6_t*, TC6Regs_Event_t event, void *pGlobalTag);

extern	int	tc6_onSpiTransaction(uint8_t tc6instance, uint8_t *pTx, uint8_t *pRx, uint16_t len, void *pGlobalTag);

extern	uint32_t	tc6regs_getTicksMs(void);

extern	void	t1s_onRawTxPacket(void *pGlobalTag, void *pTx, uint16_t len);
#endif


/** Functions adapting from uppercase to lowercase first letters.
 ** This avoids that these callback functions, which are required
 ** by the oa-tc6 library, to be exported at Go level.
 ** This also avoids some problems regarding const.
 **/

void
TC6_CB_OnNeedService(TC6_t *pInst, void *pGlobalTag) {
	tc6_onNeedService(pInst, pGlobalTag);
}

void
TC6_CB_OnError(TC6_t *pInst, TC6_Error_t err, void *pGlobalTag)
{
	tc6_onError(pInst, err, pGlobalTag);
} 

void
TC6_CB_OnRxEthernetSlice(TC6_t *pInst, const uint8_t *pRx, uint16_t offset, uint16_t len, void *pGlobalTag)
{
	tc6_onRxEthernetSlice(pInst, (void*)pRx, offset, len, pGlobalTag);
}

void
TC6_CB_OnRxEthernetPacket(TC6_t *pInst, bool success, uint16_t len, uint64_t *rxTimestamp, void *pGlobalTag)
{
	tc6_onRxEthernetPacket(pInst, success, len, castUint64 rxTimestamp, pGlobalTag);
}

void
TC6Regs_CB_OnEvent(TC6_t *pInst, TC6Regs_Event_t event, void *pTag)
{
	tc6regs_onEvent(pInst, event, pTag);	
}

bool
TC6_CB_OnSpiTransaction(uint8_t tc6instance, uint8_t *pTx, uint8_t *pRx, uint16_t len, void *pGlobalTag)
{
	return tc6_onSpiTransaction(tc6instance, pTx, pRx, len, pGlobalTag) != 0;
}

uint32_t
TC6Regs_CB_GetTicksMs(void)
{
	return tc6regs_getTicksMs();
}


/* Glue code calling TC6_SendRawEthernetPacket, and providing
 * onRawTx for calling back to Go.
 */
static void
onRawTx(TC6_t *pInst, const uint8_t *pTx, uint16_t len, void *pTag, void *pGlobalTag)
{
	t1s_onRawTxPacket(pGlobalTag, (void*)pTx, len);	
}

int
t1s_sendRawEthPacket(TC6_t *pInst, uint8_t *pTx, uint16_t len, uint8_t tsc)
{
	return TC6_SendRawEthernetPacket(pInst, pTx, len, tsc, onRawTx, 0);
}
