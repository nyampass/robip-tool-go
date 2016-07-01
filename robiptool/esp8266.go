package robiptool

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"log"
	"strconv"
	"time"
  "crypto/md5"

	"os/exec"

  "github.com/facchinm/go-serial-native"
)

type Error struct {
	msg     string
	timeout bool
}

var ErrFailedToConnect = &Error{msg: "Failed to connect ESP8266"}
var ErrTimeout = &Error{msg: "Timed out waiting for packet", timeout: true}
var ErrInvalidHeadOfPacket = &Error{msg: "Invalid head of packet"}
var ErrInvalidSLIPEscape = &Error{msg: "Invalid SLIP escape"}
var ErrResponseDoesntMatch = &Error{msg: "Response doesn't match request"}

var retChan chan []byte
var errChan chan error

func Ports() ([]string, error) {
  if infoList, err := serial.ListPorts(); err != nil {
    return nil, err

  } else {
    ports := make([]string, len(infoList))
    for i, info := range infoList {
      ports[i] = info.Name()
    }
    return ports, nil
  }
}

type UpdateProgress func(float32)

func WriteByEsptool(filepath string, port string, progressFunc UpdateProgress) error {
	cmd := exec.Command("./tool-esptool/esptool",
    "-cd", "nodemcu",
    "-cb", "115200",
    "-cp", port, "-cf", filepath)
	var out bytes.Buffer
	cmd.Stdout = &out

	if err := cmd.Start(); err != nil {
    return err
  }

  go func() {
    log.Println(out.String())
  }()

  if err := cmd.Wait(); err != nil {
    return err
  }
	fmt.Printf("in all caps: %q\n", out.String())

  return nil
}

func WriteDataToPort(filepath string, port string, progressFunc UpdateProgress) error {
  progressFunc(0)
	fmt.Printf("write: %s\n", port)
	options := serial.RawOptions
	options.Mode = serial.MODE_READ_WRITE
	options.BitRate = 115200
  // options.FlowControl = 0
	p, err := options.Open(port)
	if err != nil {
    log.Panic(err)
		return err
	}
	defer p.Close()

	if err := connect(p); err != nil {
    log.Panic(err)
		return err
	}

  log.Println("in runStub")

	if err = runStub(p, CESANTA_FLASHER_STUB); err != nil {
    log.Panic(err)
		return err
	}

  packet := <-retChan
  err = <-errChan

	if err != nil {
    log.Panic(err)
    return err

	} else if string(packet) != "OHAI" {
		return &Error{msg: fmt.Sprintf("Failed to connect to the flasher: %s", string(packet))}
	}

	if image, err := ioutil.ReadFile(filepath); err != nil {
		return err

	} else {
		buff := new(bytes.Buffer)
		if image[0] == 0xe9 {
			buff.Write(image[0:2])
			buff.Write(pack("BB", uint8(0), uint8(0)))
			buff.Write(image[4:])
		} else {
			buff.Write(image)
		}
		log.Println(len(buff.Bytes()))

		if buff.Len() % ESP_FLASH_SECTOR != 0 {
			buff.Write(bytes.Repeat([]byte{0xff}, ESP_FLASH_SECTOR-(buff.Len()%ESP_FLASH_SECTOR)))
		}

		log.Println(len(buff.Bytes()))

    log.Println("in flashWrite")

		if err := flashWrite(p, 0, buff.Bytes(), progressFunc); err != nil {
      log.Panic(err)
    }

		if err := bootFw(p); err != nil {
      log.Panic(err)
    }

    progressFunc(100)
	}

	return nil
}

const CMD_FLASH_WRITE = 1
const CMD_FLASH_READ = 2
const CMD_FLASH_DIGEST = 3
const CMD_BOOT_FW = 6

func read(p *serial.Port) ([]byte, error) {
  ret := <-retChan
  err := <-errChan

  log.Printf("read: %v, %v\n", ret, err)

  return ret, err
}

func flashWrite(p *serial.Port, addr int, data []byte, progressFunc UpdateProgress) error {
  log.Println("flashWrite")
	write(p, pack("B", uint8(CMD_FLASH_WRITE)))
	write(p, pack("III", uint32(addr), uint32(len(data)), uint32(1)))

	var numSent uint32 = 0
	var numWritten uint32 = 0

	for numWritten < uint32(len(data)) {
    log.Printf("flashWrite: %d\n", numWritten)
		if packet, err := read(p); err != nil {
      log.Printf("err: %v", err)
			return err

		} else {
      log.Printf("flashWrite: %v\n", packet)
			if len(packet) == 4 {
				numWritten = unpack("I", packet)[0]
			} else if len(packet) == 1 {
				statusCode := unpack("B", packet)[0]
				return &Error{msg: fmt.Sprintf("Write failure, status: %d", statusCode)}
			} else {
				return &Error{msg: "Unexpected packet with writing"}
			}

      percent := float32(numWritten) * 100.0 / float32(len(data))
      log.Printf("%d (%f %%)\n", numWritten, percent)
      progressFunc(percent)

			for numSent-numWritten < 5120 {
				p.Write(data[numSent:numSent+1024])
				numSent += 1024
			}
		}
	}

  log.Println("done")

	if packet, err := read(p); err != nil {
    log.Printf("err(2): %v\n", err)
		return err

	} else {
		log.Println(packet)
		if len(packet) != 16 {
      return &Error{msg: "Expected digest"}
    }
    expectedDigest := fmt.Sprintf("%x", md5.Sum(data))
		digest := hex.EncodeToString(packet)
    if digest != expectedDigest {
       return &Error{msg: "Digest mismatch"}
    }

    if packet, err = read(p); err != nil {
      return err
    } else if len(packet) != 1 {
      return &Error{msg: "Expected status"}
    } else if unpack("B", packet)[0] != 0 {
      return &Error{msg: "Write failure"}
    }
	}

  return nil
}

func bootFw(p *serial.Port) error {
  write(p, pack("B", uint8(CMD_BOOT_FW)))
  if ret, err := read(p); err != nil {
    return err

  } else if len(ret) != 1 {
    return &Error{msg: "Expected status"}

  } else if unpack("B", ret)[0] != 0 {
    return &Error{msg: "Boot failure"}
  }
  return nil
}

func connect(p *serial.Port) error {
	log.Println("Connectingg...")
  slipReader(p)
	for i := 0; i < 4; i++ {
		log.Println(i)

		p.SetDTR(serial.DTR_OFF)

    resetSlipReader()

		p.SetRTS(serial.DTR_ON)
		wait(50)

		p.SetDTR(serial.DTR_ON)
		p.SetRTS(serial.DTR_OFF)
		wait(50)

		p.SetRTS(serial.DTR_OFF)

		setTimeout(p, 300)

		for j := 0; j < 4; j++ {
			p.ResetInput()
			error := sync(p)
			p.ResetOutput()

			if error != nil {
				log.Println(error)
				wait(50)

			} else {
				log.Println("connected!")
				setTimeout(p, 5000)
				return nil
			}
		}
	}
	return ErrFailedToConnect
}

func setTimeout(p *serial.Port, milliSecond int) {
	p.SetDeadline(time.Now().Add(time.Duration(milliSecond) * time.Millisecond))
}

const (
	ESP_FLASH_BEGIN = iota + 0x02
	ESP_FLASH_DATA
	ESP_FLASH_END
	ESP_MEM_BEGIN
	ESP_MEM_END
	ESP_MEM_DATA
	ESP_SYNC
	ESP_WRITE_REG
	ESP_READ_REG
)

const ESP_CHECKSUM_MAGIC = 0xef

const ESP_FLASH_SECTOR = 0x1000

type CommandParams struct {
	op   uint8
	chk  uint32
	data []byte
}

func write(p *serial.Port, packet []byte) error {
	p.Write([]byte{0xc0})
	p.Write(bytes.Replace(bytes.Replace(packet, []byte{0xdb}, []byte{0xdb, 0xdd}, -1),
		[]byte{0xc0}, []byte{0xdb, 0xdc}, -1))
	p.Write([]byte{0xc0})

	return nil
}

func command(p *serial.Port, params CommandParams) (val uint32, body []byte, err error) {
	if params.op != 0x00 {
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.LittleEndian, uint8(0))
		binary.Write(buf, binary.LittleEndian, params.op)
		binary.Write(buf, binary.LittleEndian, uint16(len(params.data)))
		binary.Write(buf, binary.LittleEndian, params.chk)
		if len(params.data) > 0 {
			binary.Write(buf, binary.LittleEndian, params.data)
		}

		err := write(p, buf.Bytes())
		if err != nil {
			return 0, nil, err
		}
	}

	for retry := 0; retry < 100; retry++ {
		buf, err := read(p)
		if err != nil {
			return 0, nil, err
		}
		if len(buf) < 8 {
			continue
		}
		p := bytes.NewBuffer(buf)
		log.Println(buf)
		log.Println(p)
		var resp, opRet uint8
		var lenRet uint16
		var val uint32
		binary.Read(p, binary.LittleEndian, &resp)
		binary.Read(p, binary.LittleEndian, &opRet)
		binary.Read(p, binary.LittleEndian, &lenRet)
		binary.Read(p, binary.LittleEndian, &val)

		log.Println("command resopnses:")
		log.Println(resp)
		log.Println(opRet)
		log.Println(lenRet)
		log.Println(val)

		var body []byte = buf[8:]

		if params.op == 0x00 || opRet == params.op {
			log.Println("command: response: ")
			log.Println(val)
			log.Println(body)
			return val, body, nil
		}
	}

	return 0, nil, ErrResponseDoesntMatch
}

func unpack(format string, b []byte) []uint32 {
	buffer := bytes.NewBuffer(b)
	var vals []uint32 = make([]uint32, len(format))

	for i, f := range format {
		switch f {
		case 'B':
			var v uint8
			binary.Read(buffer, binary.LittleEndian, &v)
			vals[i] = uint32(v)

		case 'H':
			var v uint16
			binary.Read(buffer, binary.LittleEndian, &v)
			vals[i] = uint32(v)

		case 'I':
			var v uint32
			binary.Read(buffer, binary.LittleEndian, &v)
			vals[i] = v

		default:
			log.Panic("unknown format")
		}
	}

	return vals
}

func sync(p *serial.Port) error {
	log.Println("sync")

	var syncData bytes.Buffer
	syncData.Write([]byte{0x07, 0x07, 0x12, 0x20})
	syncData.Write(bytes.Repeat([]byte{0x55}, 32))

	_, _, err := command(p, CommandParams{op: ESP_SYNC, data: syncData.Bytes()})
	if err != nil {
		return err
	}

	for i := 0; i < 7; i++ {
		command(p, CommandParams{})
	}

	return nil
}

func wait(milliSecond int) {
	time.Sleep(time.Duration(milliSecond) * time.Millisecond)
}

var partialPacket bytes.Buffer
var isInitializedPartialPacket = false
var inEscape = false

func resetSlipReader() {
  partialPacket.Reset()
  isInitializedPartialPacket = false
  inEscape = false
}

func slipReader(p *serial.Port) () {
	fmt.Println("in slip reader")

  retChan = make(chan []byte, 0)
  errChan = make(chan error, 0)

  go func() {
    for {
      log.Println("slip reader: for")
      waiting, _ := p.InputWaiting()
      byteSize := waiting
      if byteSize == 0 {
        byteSize = 1
      }
      log.Printf("byteSize: %d, waiting: %d\n", byteSize, waiting)
      buf := make([]byte, byteSize)

      setTimeout(p, 300)
      if count, err := p.Read(buf); err != nil {
        retChan<- nil
        errChan<- err

      } else if count == 0 {
        retChan<- nil
        errChan<- ErrTimeout

      } else {
        log.Println("slip reader: else")
        log.Println(buf)
        for _, b := range buf {
          if !isInitializedPartialPacket {
            if b == 0xc0 {
              isInitializedPartialPacket = true
              partialPacket.Reset()

            } else {
              retChan<- nil
              errChan<- ErrInvalidHeadOfPacket
            }

          } else if inEscape {
            inEscape = false
            if b == 0xdc {
              partialPacket.WriteByte(0xc0)

            } else if b == 0xdd {
              partialPacket.WriteByte(0xdb)

            } else {
              retChan<- nil
              errChan<- ErrInvalidSLIPEscape
            }

          } else if b == 0xdb {
            inEscape = true

          } else if b == 0xc0 {
            bytes := partialPacket.Bytes()
            sendBytes := make([]byte, len(bytes))
            copy(sendBytes, bytes)
            log.Printf("slipReader: %v, %v, %v, %v\n", b,  buf, bytes, sendBytes)

            retChan<- sendBytes
            errChan<- nil

            isInitializedPartialPacket = false

          } else {
            partialPacket.WriteByte(b)
          }
        }
      }
    }
  }()
}

func hexint(s string) uint8 {
	n, err := strconv.ParseUint(s, 16, 32)
	if err != nil {
		panic(err)
	}
	return uint8(n)
}

func pack(format string, vals ...interface{}) []byte {
	var data = new(bytes.Buffer)
	for _, v := range vals {
		switch v := v.(type) {
		case uint8, uint32, uint64:
			binary.Write(data, binary.LittleEndian, v)
		default:
			fmt.Println("unknown")
		}
	}
	return data.Bytes()
}

func bytes2uint32(bytes []byte) uint32 {
	var val uint32 = 0
	for _, b := range bytes {
		val = val<<8 + uint32(b)
	}
	return val
}

func checksum(data []byte, state uint32) uint32 {
	for _, b := range data {
		state ^= uint32(b)
	}
	log.Printf("checksum:%d\n", state)

	return state
}

func checksumMagicState(data []byte) uint32 {
	return checksum(data, ESP_CHECKSUM_MAGIC)
}

func memBegin(p *serial.Port, size int, blocks int, blockSize int, offset int) error {
	log.Printf("mem_begin: %d, %d, %d, %d\n", size, blocks, blockSize, offset)
	_, body, err := command(p,
		CommandParams{op: ESP_MEM_BEGIN,
			data: pack("IIII", uint32(size), uint32(blocks), uint32(blockSize), uint32(offset))})

	if err != nil {
		return err
	}

	if bytes2uint32(body) != 0 {
		return &Error{msg: "Failed to enter RAM download mode"}
	}

	return nil
}

func memBlock(p *serial.Port, data []byte, seq int) error {
	buf := new(bytes.Buffer)
	buf.Write(pack("IIII", uint32(len(data)), uint32(seq), uint32(0), uint32(0)))
	buf.Write(data)
	_, body, err := command(p, CommandParams{op: ESP_MEM_DATA, data: buf.Bytes(), chk: checksumMagicState(data)})
	if err != nil {
		return err
	}

	if bytes2uint32(body) != 0 {
		return &Error{msg: "Failed to write to target RAM"}
	}

	return nil
}

func memFinish(p *serial.Port, entrypoint int) error {
	var isEntrypointZero uint32 = 0
	if entrypoint == 0 {
		isEntrypointZero = 1
	}
	if _, body, err := command(p, CommandParams{op: ESP_MEM_END,
		data: pack("II", isEntrypointZero, uint32(entrypoint))}); err != nil {
		return err
	} else {
		if bytes2uint32(body) != 0 {
			return &Error{msg: "Failed to write to target RAM"}
		}
		return nil
	}
}

func runStub(p *serial.Port, stub Stub) error {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, uint32(0))
	for i := 0; i < len(stub.code); i += 2 {
		binary.Write(buf, binary.LittleEndian, hexint(stub.code[i:i+2]))
	}

	var bytes = buf.Bytes()
	if err := memBegin(p, len(bytes), 1, len(bytes), stub.paramsStart); err != nil {
		return err
	}

	if err := memBlock(p, bytes, 0); err != nil {
		return err
	}

	buf.Reset()
	for i := 0; i < len(stub.data); i += 2 {
		binary.Write(buf, binary.LittleEndian, hexint(stub.data[i:i+2]))
	}
	bytes = buf.Bytes()

	if err := memBegin(p, len(bytes), 1, len(bytes), stub.dataStart); err != nil {
		return err
	}

	if err := memBlock(p, bytes, 0); err != nil {
		return err
	}

	if err := memFinish(p, stub.entry); err != nil {
		return err
	}

	return nil
}

func (e *Error) Error() string {
	return e.msg
}

func (e *Error) Timeout() bool {
	return e.timeout
}

type Stub struct {
	codeStart, entry, numParams, paramsStart, dataStart int
	code, data                                          string
}

var CESANTA_FLASHER_STUB = Stub{codeStart: 1074790404, code: "080000601C000060000000601000006031FCFF71FCFF81FCFFC02000680332D218C020004807404074DCC48608005823C0200098081BA5A92392450058031B555903582337350129230B446604DFC6F3FF21EEFFC0200069020DF0000000010078480040004A0040B449004012C1F0C921D911E901DD0209312020B4ED033C2C56C2073020B43C3C56420701F5FFC000003C4C569206CD0EEADD860300202C4101F1FFC0000056A204C2DCF0C02DC0CC6CCAE2D1EAFF0606002030F456D3FD86FBFF00002020F501E8FFC00000EC82D0CCC0C02EC0C73DEB2ADC460300202C4101E1FFC00000DC42C2DCF0C02DC056BCFEC602003C5C8601003C6C4600003C7C08312D0CD811C821E80112C1100DF0000C180000140010400C0000607418000064180000801800008C18000084180000881800009018000018980040880F0040A80F0040349800404C4A0040740F0040800F0040980F00400099004012C1E091F5FFC961CD0221EFFFE941F9310971D9519011C01A223902E2D1180C02226E1D21E4FF31E9FF2AF11A332D0F42630001EAFFC00000C030B43C2256A31621E1FF1A2228022030B43C3256B31501ADFFC00000DD023C4256ED1431D6FF4D010C52D90E192E126E0101DDFFC0000021D2FF32A101C020004802303420C0200039022C0201D7FFC00000463300000031CDFF1A333803D023C03199FF27B31ADC7F31CBFF1A3328030198FFC0000056C20E2193FF2ADD060E000031C6FF1A3328030191FFC0000056820DD2DD10460800000021BEFF1A2228029CE231BCFFC020F51A33290331BBFFC02C411A332903C0F0F4222E1D22D204273D9332A3FFC02000280E27B3F721ABFF381E1A2242A40001B5FFC00000381E2D0C42A40001B3FFC0000056120801B2FFC00000C02000280EC2DC0422D2FCC02000290E01ADFFC00000222E1D22D204226E1D281E22D204E7B204291E860000126E012198FF32A0042A21C54C003198FF222E1D1A33380337B202C6D6FF2C02019FFFC000002191FF318CFF1A223A31019CFFC00000218DFF1C031A22C549000C02060300003C528601003C624600003C72918BFF9A110871C861D851E841F83112C1200DF00010000068100000581000007010000074100000781000007C100000801000001C4B0040803C004091FDFF12C1E061F7FFC961E941F9310971D9519011C01A66290621F3FFC2D1101A22390231F2FF0C0F1A33590331EAFFF26C1AED045C2247B3028636002D0C016DFFC0000021E5FF41EAFF2A611A4469040622000021E4FF1A222802F0D2C0D7BE01DD0E31E0FF4D0D1A3328033D0101E2FFC00000561209D03D2010212001DFFFC000004D0D2D0C3D01015DFFC0000041D5FFDAFF1A444804D0648041D2FF1A4462640061D1FF106680622600673F1331D0FF10338028030C43853A002642164613000041CAFF222C1A1A444804202FC047328006F6FF222C1A273F3861C2FF222C1A1A6668066732B921BDFF3D0C1022800148FFC0000021BAFF1C031A2201BFFFC000000C024603005C3206020000005C424600005C5291B7FF9A110871C861D851E841F83112C1200DF0B0100000C0100000D010000012C1E091FEFFC961D951E9410971F931CD039011C0ED02DD0431A1FF9C1422A06247B302062D0021F4FF1A22490286010021F1FF1A223902219CFF2AF12D0F011FFFC00000461C0022D110011CFFC0000021E9FFFD0C1A222802C7B20621E6FF1A22F8022D0E3D014D0F0195FFC000008C5222A063C6180000218BFF3D01102280F04F200111FFC00000AC7D22D1103D014D0F010DFFC0000021D6FF32D110102280010EFFC0000021D3FF1C031A220185FFC00000FAEEF0CCC056ACF821CDFF317AFF1A223A310105FFC0000021C9FF1C031A22017CFFC000002D0C91C8FF9A110871C861D851E841F83112C1200DF0000200600000001040020060FFFFFF0012C1E00C02290131FAFF21FAFF026107C961C02000226300C02000C80320CC10564CFF21F5FFC02000380221F4FF20231029010C432D010163FFC0000008712D0CC86112C1200DF00080FE3F8449004012C1D0C9A109B17CFC22C1110C13C51C00261202463000220111C24110B68202462B0031F5FF3022A02802A002002D011C03851A0066820A280132210105A6FF0607003C12C60500000010212032A01085180066A20F2221003811482105B3FF224110861A004C1206FDFF2D011C03C5160066B20E280138114821583185CFFF06F7FF005C1286F5FF0010212032A01085140066A20D2221003811482105E1FF06EFFF0022A06146EDFF45F0FFC6EBFF000001D2FFC0000006E9FF000C022241100C1322C110C50F00220111060600000022C1100C13C50E0022011132C2FA303074B6230206C8FF08B1C8A112C1300DF0000000000010404F484149007519031027000000110040A8100040BC0F0040583F0040CC2E00401CE20040D83900408000004021F4FF12C1E0C961C80221F2FF097129010C02D951C91101F4FFC0000001F3FFC00000AC2C22A3E801F2FFC0000021EAFFC031412A233D0C01EFFFC000003D0222A00001EDFFC00000C1E4FF2D0C01E8FFC000002D0132A004450400C5E7FFDD022D0C01E3FFC00000666D1F4B2131DCFF4600004B22C0200048023794F531D9FFC0200039023DF08601000001DCFFC000000871C861D85112C1200DF000000012C1F002610301EAFEC00000083112C1100DF000643B004012C1D0E98109B1C9A1D991F97129013911E2A0C001FAFFC00000CD02E792F40C0DE2A0C0F2A0DB860D00000001F4FFC00000204220E71240F7921C22610201EFFFC0000052A0DC482157120952A0DD571205460500004D0C3801DA234242001BDD3811379DC5C6000000000C0DC2A0C001E3FFC00000C792F608B12D0DC8A1D891E881F87112C1300DF00000", entry: 1074792180, numParams: 1, paramsStart: 1074790400, data: "FE0510401A0610403B0610405A0610407A061040820610408C0610408C061040", dataStart: 1073643520}
