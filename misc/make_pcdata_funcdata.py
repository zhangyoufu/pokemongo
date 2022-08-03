functab_ea = ida_name.get_name_ea(idaapi.BADADDR, 'functab')
pctab_ea = ida_name.get_name_ea(idaapi.BADADDR, 'pctab')
funcdata_ea = ida_name.get_name_ea(idaapi.BADADDR, 'go.func._ptr_')

def handle_func_struct(ea):
	npcdata = ida_bytes.get_dword(ea + 0x1C)
	nfuncdata = ida_bytes.get_byte(ea + 0x27)
	ea += 0x28
	assert npcdata <= 4 and nfuncdata <= 8
	for i in range(npcdata):
		if ida_bytes.create_dword(ea, 4):
			ida_offset.op_offset(ea, 0, ida_nalt.REF_OFF32, idaapi.BADADDR, pctab_ea)
		ea += 4
	for i in range(nfuncdata):
		if ida_bytes.create_dword(ea, 4) and ida_bytes.get_dword(ea) != 0xFFFFFFFF:
			ida_offset.op_offset(ea, 0, ida_nalt.REF_OFF32, idaapi.BADADDR, funcdata_ea)
		ea += 4

def main():
	lastpc = seg_text.size()
	ea = functab_ea
	while 1:
		func_pc_off = ida_bytes.get_dword(ea)
		if func_pc_off >= lastpc: break
		func_struct_ea = functab_ea + ida_bytes.get_dword(ea+4)
		handle_func_struct(func_struct_ea)
		ea += 8
	print('DONE')

if __name__ == '__main__':
	main()
