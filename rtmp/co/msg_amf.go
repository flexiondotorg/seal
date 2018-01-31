package co

import (
	"UtilsTools/identify_panic"
	"fmt"
	"log"
	"seal/rtmp/pt"
)

func (rc *RtmpConn) msgAmf(msg *pt.Message) (err error) {
	defer func() {
		if err := recover(); err != nil {
			log.Println(err, ",panic at ", identify_panic.IdentifyPanic())
		}
	}()

	var offset uint32

	// skip 1bytes to decode the amf3 command.
	if pt.RTMP_MSG_AMF3CommandMessage == msg.Header.MessageType && msg.Header.PayloadLength >= 1 {
		offset++
	}

	//read the command name.
	err, command := pt.Amf0ReadString(msg.Payload, &offset)
	if err != nil {
		return
	}

	if 0 == len(command) {
		err = fmt.Errorf("Amf0ReadString failed, command is nil.csid=%d", msg.Header.PerferCsid)
		return
	}

	log.Println("amf0/3 command or amf0/3 data, msg typeid=", msg.Header.MessageType, ",command =", command)

	switch command {
	case pt.RTMP_AMF0_COMMAND_RESULT, pt.RTMP_AMF0_COMMAND_ERROR:
		err = rc.amf0ResultError(msg)
	case pt.RTMP_AMF0_COMMAND_CONNECT:
		err = rc.amf0Connect(msg)
	case pt.RTMP_AMF0_COMMAND_CREATE_STREAM:
		err = rc.amf0CreateStream(msg)
	case pt.RTMP_AMF0_COMMAND_PLAY:
		err = rc.amf0Play(msg)
	case pt.RTMP_AMF0_COMMAND_PAUSE:
		err = rc.amf0Pause(msg)
	case pt.RTMP_AMF0_COMMAND_RELEASE_STREAM:
		err = rc.amf0ReleaseStream(msg)
	case pt.RTMP_AMF0_COMMAND_FC_PUBLISH:
		err = rc.amf0FcPublish(msg)
	case pt.RTMP_AMF0_COMMAND_PUBLISH:
		err = rc.amf0Publish(msg)
	case pt.RTMP_AMF0_COMMAND_UNPUBLISH:
		err = rc.amf0UnPublish(msg)
	case pt.RTMP_AMF0_COMMAND_KEEPLIVE:
		//todo.
	case pt.RTMP_AMF0_COMMAND_ENABLEVIDEO:
		//todo.
	case pt.RTMP_AMF0_DATA_SET_DATAFRAME, pt.RTMP_AMF0_DATA_ON_METADATA:
		err = rc.amf0Meta(msg)
	case pt.RTMP_AMF0_DATA_ON_CUSTOMDATA:
		err = rc.amf0OnCustom(msg)
	case pt.RTMP_AMF0_COMMAND_CLOSE_STREAM:
		err = rc.amf0CloseStream(msg)
	case pt.RTMP_AMF0_COMMAND_ON_BW_DONE:
		err = rc.amf0OnBwDone(msg)
	case pt.RTMP_AMF0_COMMAND_ON_STATUS:
		err = rc.amf0OnStatus(msg)
	case pt.RTMP_AMF0_COMMAND_GET_STREAM_LENGTH:
		err = rc.amf0GetStreamLen(msg)
	case pt.RTMP_AMF0_COMMAND_INSERT_KEYFREAME:
		//todo
	case pt.RTMP_AMF0_DATA_SAMPLE_ACCESS:
		err = rc.amf0SampleAccess(msg)
	default:
		log.Println("msg amf unknown command name=", command)
	}

	if err != nil {
		return
	}
	return
}

func (rc *RtmpConn) amf0ResultError(msg *pt.Message) (err error) {
	defer func() {
		if err := recover(); err != nil {
			log.Println(err, ",panic at ", identify_panic.IdentifyPanic())
		}
	}()

	log.Println("Amf0ResultError")

	var offset uint32

	var transaction_id float64
	err, transaction_id = pt.Amf0ReadNumber(msg.Payload, &offset)
	if err != nil {
		log.Println("read transaction id failed when decode msg.")
		return
	}

	req_command_name := rc.Requests[transaction_id]
	if 0 == len(req_command_name) {
		err = fmt.Errorf("can not find request command name.")
		return
	}
	switch req_command_name {
	case pt.RTMP_AMF0_COMMAND_CONNECT:
		p := pt.ConnectResPacket{}
		err = p.Decode(msg.Payload)
	case pt.RTMP_AMF0_COMMAND_CREATE_STREAM:
		p := pt.CreateStreamResPacket{}
		err = p.Decode(msg.Payload)
	case pt.RTMP_AMF0_COMMAND_RELEASE_STREAM,
		pt.RTMP_AMF0_COMMAND_FC_PUBLISH,
		pt.RTMP_AMF0_COMMAND_UNPUBLISH:
		p := pt.FmleStartResPacket{}
		err = p.Decode(msg.Payload)
	default:
		log.Println("result/error: unknown request command name=", req_command_name)
	}

	if err != nil {
		log.Println("decode result or error response msg failed.")
		return
	}

	return
}

func (rc *RtmpConn) amf0Connect(msg *pt.Message) (err error) {
	defer func() {
		if err := recover(); err != nil {
			log.Println(err, ",panic at ", identify_panic.IdentifyPanic())
		}
	}()

	log.Println("Amf0Connect")

	p := pt.ConnectPacket{}
	err = p.Decode(msg.Payload)
	if err != nil {
		log.Println("decode conncet pkt faile.err=", err)
		return
	}

	if nil == p.GetObjectProperty("tcUrl") {
		err = fmt.Errorf("no tcUrl info in connect.")
		return
	}
	rc.ConnectInfo.TcUrl = p.GetObjectProperty("tcUrl").(string)
	if o := p.GetObjectProperty("pageUrl"); o != nil {
		rc.ConnectInfo.PageUrl = o.(string)
	}
	if o := p.GetObjectProperty("swfUrl"); o != nil {
		rc.ConnectInfo.SwfUrl = o.(string)
	}
	if o := p.GetObjectProperty("objectEncoding"); o != nil {
		rc.ConnectInfo.ObjectEncoding = o.(float64)
	}

	log.Println("decode connect pkt success.", p, ", info=", rc.ConnectInfo)

	var pkt pt.ConnectResPacket

	pkt.CommandName = pt.RTMP_AMF0_COMMAND_RESULT
	pkt.TransactionId = 1

	pkt.Props = append(pkt.Props, pt.Amf0Object{
		PropertyName: "fmsVer",
		Value:        "FMS/" + pt.FMS_VERSION,
		ValueType:    pt.RTMP_AMF0_String,
	})

	pkt.Props = append(pkt.Props, pt.Amf0Object{
		PropertyName: "capabilities",
		Value:        127.0,
		ValueType:    pt.RTMP_AMF0_Number,
	})

	pkt.Props = append(pkt.Props, pt.Amf0Object{
		PropertyName: "mode",
		Value:        1.0,
		ValueType:    pt.RTMP_AMF0_Number,
	})

	pkt.Props = append(pkt.Props, pt.Amf0Object{
		PropertyName: pt.StatusLevel,
		Value:        pt.StatusLevelStatus,
		ValueType:    pt.RTMP_AMF0_String,
	})

	pkt.Props = append(pkt.Props, pt.Amf0Object{
		PropertyName: pt.StatusCode,
		Value:        pt.StatusCodeConnectSuccess,
		ValueType:    pt.RTMP_AMF0_String,
	})

	pkt.Props = append(pkt.Props, pt.Amf0Object{
		PropertyName: pt.StatusDescription,
		Value:        "Connection succeeded",
		ValueType:    pt.RTMP_AMF0_String,
	})

	pkt.Props = append(pkt.Props, pt.Amf0Object{
		PropertyName: "objectEncoding",
		Value:        rc.ConnectInfo.ObjectEncoding,
		ValueType:    pt.RTMP_AMF0_Number,
	})

	pkt.Info = append(pkt.Info, pt.Amf0Object{
		PropertyName: "version",
		Value:        pt.FMS_VERSION,
		ValueType:    pt.RTMP_AMF0_String,
	})

	pkt.Info = append(pkt.Info, pt.Amf0Object{
		PropertyName: "seal_license",
		Value:        "The MIT License (MIT)",
		ValueType:    pt.RTMP_AMF0_String,
	})

	pkt.Info = append(pkt.Info, pt.Amf0Object{
		PropertyName: "seal_authors",
		Value:        "YangKai",
		ValueType:    pt.RTMP_AMF0_String,
	})

	pkt.Info = append(pkt.Info, pt.Amf0Object{
		PropertyName: "seal_email",
		Value:        "beyondyangkai@gmail.com",
		ValueType:    pt.RTMP_AMF0_String,
	})

	pkt.Info = append(pkt.Info, pt.Amf0Object{
		PropertyName: "seal_copyright",
		Value:        "Copyright (c) 2018 YangKai",
		ValueType:    pt.RTMP_AMF0_String,
	})

	pkt.Info = append(pkt.Info, pt.Amf0Object{
		PropertyName: "seal_sig",
		Value:        "seal",
		ValueType:    pt.RTMP_AMF0_String,
	})

	err = rc.SendPacket(&pkt, 0)
	if err != nil {
		log.Println("response connect error.", err)
		return
	}

	log.Println("send connect response success.")

	return
}

func (rc *RtmpConn) amf0CreateStream(msg *pt.Message) (err error) {
	defer func() {
		if err := recover(); err != nil {
			log.Println(err, ",panic at ", identify_panic.IdentifyPanic())
		}
	}()

	log.Println("Amf0CreateStream")

	p := pt.CreateStreamPacket{}
	err = p.Decode(msg.Payload)
	if nil != err {
		log.Println("decode create stream failed.")
		return
	}

	//createStream response
	var pkt pt.CreateStreamResPacket
	pkt.CommandName = pt.RTMP_AMF0_COMMAND_RESULT
	pkt.TransactionId = p.TransactionId
	pkt.StreamId = rc.DefaultStreamId

	err = rc.SendPacket(&pkt, 0)
	if err != nil {
		log.Println("send createStream response failed. err=", err)
		return
	}
	log.Println("send createStream response success.")

	return
}

func (rc *RtmpConn) amf0Play(msg *pt.Message) (err error) {
	defer func() {
		if err := recover(); err != nil {
			log.Println(err, ",panic at ", identify_panic.IdentifyPanic())
		}
	}()

	log.Println("Amf0Play")

	p := pt.PlayPacket{}
	err = p.Decode(msg.Payload)
	if err != nil {
		return
	}

	// StreamBegin
	if true {
		var pp pt.UserControlPacket
		pp.EventType = pt.SrcPCUCStreamBegin
		pp.EventData = msg.Header.StreamId
		err = rc.SendPacket(&pp, 0)
		if err != nil {
			return
		}
	}

	// onStatus(NetStream.Play.Reset)
	if true {
		var pp pt.OnStatusCallPacket
		pp.CommandName = pt.RTMP_AMF0_COMMAND_ON_STATUS
		pp.Data = append(pp.Data, pt.Amf0Object{
			PropertyName: pt.StatusLevel,
			Value:        pt.StatusLevelStatus,
			ValueType:    pt.RTMP_AMF0_String,
		})

		pp.Data = append(pp.Data, pt.Amf0Object{
			PropertyName: pt.StatusCode,
			Value:        pt.StatusCodeStreamReset,
			ValueType:    pt.RTMP_AMF0_String,
		})

		pp.Data = append(pp.Data, pt.Amf0Object{
			PropertyName: pt.StatusDescription,
			Value:        "Playing and resetting stream.",
			ValueType:    pt.RTMP_AMF0_String,
		})
		pp.Data = append(pp.Data, pt.Amf0Object{
			PropertyName: pt.StatusDetails,
			Value:        "stream",
			ValueType:    pt.RTMP_AMF0_String,
		})
		pp.Data = append(pp.Data, pt.Amf0Object{
			PropertyName: pt.StatusClientId,
			Value:        pt.RTMP_SIG_CLIENT_ID,
			ValueType:    pt.RTMP_AMF0_String,
		})
		err = rc.SendPacket(&pp, uint32(rc.DefaultStreamId))
		if err != nil {
			return
		}

	}

	// onStatus(NetStream.Play.Start)
	if true {
		var pp pt.OnStatusCallPacket
		pp.CommandName = pt.RTMP_AMF0_COMMAND_ON_STATUS

		pp.Data = append(pp.Data, pt.Amf0Object{
			PropertyName: pt.StatusLevel,
			Value:        pt.StatusLevelStatus,
			ValueType:    pt.RTMP_AMF0_String,
		})

		pp.Data = append(pp.Data, pt.Amf0Object{
			PropertyName: pt.StatusCode,
			Value:        pt.StatusCodeStreamStart,
			ValueType:    pt.RTMP_AMF0_String,
		})

		pp.Data = append(pp.Data, pt.Amf0Object{
			PropertyName: pt.StatusDescription,
			Value:        "Started playing stream.",
			ValueType:    pt.RTMP_AMF0_String,
		})

		pp.Data = append(pp.Data, pt.Amf0Object{
			PropertyName: pt.StatusDetails,
			Value:        "stream",
			ValueType:    pt.RTMP_AMF0_String,
		})

		pp.Data = append(pp.Data, pt.Amf0Object{
			PropertyName: pt.StatusClientId,
			Value:        pt.RTMP_SIG_CLIENT_ID,
			ValueType:    pt.RTMP_AMF0_String,
		})
		err = rc.SendPacket(&pp, uint32(rc.DefaultStreamId))
		if err != nil {
			return
		}
		log.Println("send NetStream.Play.Reset response success.")
	}

	// |RtmpSampleAccess(false, false)
	if true {
		var pp pt.SampleAccessPacket
		pp.CommandName = pt.RTMP_AMF0_DATA_SAMPLE_ACCESS
		pp.AudioSampleAccess = true
		pp.VideoSampleAccess = true
		err = rc.SendPacket(&pp, uint32(rc.DefaultStreamId))
		if err != nil {
			return
		}
		log.Println("send RtmpSampleAccess success")
	}

	// onStatus(NetStream.Data.Start)
	if true {
		var pp pt.OnStatusDataPacket
		pp.CommandName = pt.RTMP_AMF0_COMMAND_ON_STATUS
		pp.Data = append(pp.Data, pt.Amf0Object{
			PropertyName: pt.StatusLevel,
			Value:        pt.StatusLevelStatus,
			ValueType:    pt.RTMP_AMF0_String,
		})

		pp.Data = append(pp.Data, pt.Amf0Object{
			PropertyName: pt.StatusCode,
			Value:        pt.StatusCodeDataStart,
			ValueType:    pt.RTMP_AMF0_String,
		})

		pp.Data = append(pp.Data, pt.Amf0Object{
			PropertyName: pt.StatusDescription,
			Value:        "Started playing stream data.",
			ValueType:    pt.RTMP_AMF0_String,
		})
		err = rc.SendPacket(&pp, uint32(rc.DefaultStreamId))
		if err != nil {
			return
		}
		log.Println("send NetStream.Data.Start success.")
	}

	return
}

func (rc *RtmpConn) amf0Pause(msg *pt.Message) (err error) {
	defer func() {
		if err := recover(); err != nil {
			log.Println(err, ",panic at ", identify_panic.IdentifyPanic())
		}
	}()

	log.Println("Amf0Pause")

	p := pt.PausePacket{}
	err = p.Decode(msg.Payload)
	if err != nil {
		return
	}

	return
}

func (rc *RtmpConn) amf0ReleaseStream(msg *pt.Message) (err error) {
	defer func() {
		if err := recover(); err != nil {
			log.Println(err, ",panic at ", identify_panic.IdentifyPanic())
		}
	}()

	log.Println("Amf0ReleaseStream")

	p := pt.FmleStartPacket{}
	err = p.Decode(msg.Payload)
	if err != nil {
		return
	}

	return
}

func (rc *RtmpConn) amf0FcPublish(msg *pt.Message) (err error) {
	defer func() {
		if err := recover(); err != nil {
			log.Println(err, ",panic at ", identify_panic.IdentifyPanic())
		}
	}()

	log.Println("Amf0FcPublish")

	p := pt.FmleStartPacket{}
	err = p.Decode(msg.Payload)
	if err != nil {
		return
	}

	var pp pt.FmleStartResPacket
	pp.CommandName = pt.RTMP_AMF0_COMMAND_RESULT
	pp.TransactionId = p.TransactionId
	err = rc.SendPacket(&pp, 0)
	if err != nil {
		return
	}
	log.Println("send FcPublish response success.")

	return
}

func (rc *RtmpConn) amf0Publish(msg *pt.Message) (err error) {
	defer func() {
		if err := recover(); err != nil {
			log.Println(err, ",panic at ", identify_panic.IdentifyPanic())
		}
	}()

	log.Println("Amf0Publish")

	p := pt.PublishPacket{}
	err = p.Decode(msg.Payload)
	if err != nil {
		return
	}

	if !rc.CheckStreamCanPublish(rc.StreamName) {
		err = fmt.Errorf("stream=%s can not publish, has publishing now", rc.StreamName)
		return
	}

	rc.Role = RtmpRoleFMLEPublisher
	log.Println("a new publisher come in, stream name=", rc.StreamName)

	return
}

func (rc *RtmpConn) amf0UnPublish(msg *pt.Message) (err error) {
	defer func() {
		if err := recover(); err != nil {
			log.Println(err, ",panic at ", identify_panic.IdentifyPanic())
		}
	}()

	log.Println("Amf0UnPublish")

	p := pt.FmleStartPacket{}
	err = p.Decode(msg.Payload)
	if err != nil {
		return
	}

	var pp pt.FmleStartResPacket
	pp.CommandName = pt.RTMP_AMF0_COMMAND_RESULT
	pp.TransactionId = p.TransactionId
	err = rc.SendPacket(&pp, 0)
	if err != nil {
		return
	}
	log.Println("send unpublish response success.")

	return
}

func (rc *RtmpConn) amf0Meta(msg *pt.Message) (err error) {
	defer func() {
		if err := recover(); err != nil {
			log.Println(err, ",panic at ", identify_panic.IdentifyPanic())
		}
	}()

	log.Println("Amf0Meta")

	p := pt.OnMetaDataPacket{}
	err = p.Decode(msg.Payload)
	if err != nil {
		return
	}
	log.Println("decode meta data success, meta=", p)

	return
}

func (rc *RtmpConn) amf0OnCustom(msg *pt.Message) (err error) {
	defer func() {
		if err := recover(); err != nil {
			log.Println(err, ",panic at ", identify_panic.IdentifyPanic())
		}
	}()

	log.Println("Amf0OnCustom")

	p := pt.OnCustomDataPakcet{}
	err = p.Decode(msg.Payload)
	if err != nil {
		return
	}

	return
}

func (rc *RtmpConn) amf0CloseStream(msg *pt.Message) (err error) {
	defer func() {
		if err := recover(); err != nil {
			log.Println(err, ",panic at ", identify_panic.IdentifyPanic())
		}
	}()

	log.Println("Amf0CloseStream")

	p := pt.CloseStreamPacket{}
	err = p.Decode(msg.Payload)
	if err != nil {
		return
	}

	return
}

func (rc *RtmpConn) amf0OnBwDone(msg *pt.Message) (err error) {
	defer func() {
		if err := recover(); err != nil {
			log.Println(err, ",panic at ", identify_panic.IdentifyPanic())
		}
	}()

	log.Println("Amf0OnBwDone")

	return
}

func (rc *RtmpConn) amf0OnStatus(msg *pt.Message) (err error) {
	defer func() {
		if err := recover(); err != nil {
			log.Println(err, ",panic at ", identify_panic.IdentifyPanic())
		}
	}()

	log.Println("Amf0Onstats")

	return
}

func (rc *RtmpConn) amf0GetStreamLen(msg *pt.Message) (err error) {
	defer func() {
		if err := recover(); err != nil {
			log.Println(err, ",panic at ", identify_panic.IdentifyPanic())
		}
	}()

	log.Println("Amf0GetStreamLen")

	return
}

func (rc *RtmpConn) amf0SampleAccess(msg *pt.Message) (err error) {
	defer func() {
		if err := recover(); err != nil {
			log.Println(err, ",panic at ", identify_panic.IdentifyPanic())
		}
	}()

	log.Println("Amf0SampleAccess")

	return
}