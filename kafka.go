package loges

import (
	kafka "github.com/araddon/kafka/clients/gokafka"
	log "github.com/ngmoco/timber"
	"strconv"
	"strings"
)

// Kafka formatter, for parsing kafka messages
func KafkaFormatter(e *LineEvent) *Event {
	//2012-11-22 05:07:51 +0000 lio.home.ubuntu.log.collect.log.vm2: {"message":"runtime error: close of closed channel"}
	if ml := strings.SplitN(string(e.Data), ": ", 2); len(ml) > 1 {
		log.Debug("%v\n", strings.Join(ml, "||"))
		if len(ml[0]) > 26 {
			//d := ml[0][0:25]
			src := ml[0][26:]
			return NewEvent("golog", src, ml[1])
		}
	}
	return nil
}

func RunKafkaConsumer(msgChan chan *LineEvent, partitionstr, topic, kafkaHost string, offset, maxMsgCt uint64, maxSize uint) {
	var broker *kafka.BrokerConsumer

	parts := strings.Split(partitionstr, ",")
	if len(parts) > 1 {
		tps := kafka.NewTopicPartitions(topic, partitionstr, offset, uint32(maxSize))
		broker = kafka.NewMultiConsumer(kafkaHost, tps)
	} else {
		partition, _ := strconv.Atoi(partitionstr)
		broker = kafka.NewBrokerConsumer(kafkaHost, topic, partition, offset, uint32(maxSize))
	}

	var msgCt int
	done := make(chan bool, 1)
	kafkaMsgs := make(chan *kafka.Message)
	go broker.ConsumeOnChannel(kafkaMsgs, 1000, done)
	for msg := range kafkaMsgs {
		if msg != nil {
			msgCt++
			if uint64(msgCt) > maxMsgCt {
				panic("ending")
			}
			//msg.Print()
			msgChan <- &LineEvent{Data: msg.Payload(), Offset: msg.Offset(), Item: msg}
		} else {
			log.Error("No kafka message?")
			break
		}
	}
}
