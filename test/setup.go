package main

import (
	"fmt"
	"time"

	"github.com/rmoff/ksqldb-go"
)

func setup() (*ksqldb.Client, error) {
	//create ksqldb client
	client := ksqldb.NewClient(ksqlDBServer).Debug()
	// Create the dummy data connector
	if err := client.Execute(`
		CREATE SOURCE CONNECTOR DOGS WITH (
		'connector.class'             = 'io.mdrogalis.voluble.VolubleSourceConnector',
		'key.converter'               = 'org.apache.kafka.connect.storage.StringConverter',
		'value.converter'             = 'org.apache.kafka.connect.json.JsonConverter',
		'value.converter.schemas.enable'='false',
		'genkp.dogs.with'             = '#{Internet.uuid}',
		'genv.dogs.name.with'       = '#{Dog.name}',
		'genv.dogs.dogsize.with'       = '#{Dog.size}',
		'genv.dogs.age.with'        = '#{Dog.age}',
		'topic.dogs.throttle.ms'    = 1000 
		);
		`); err != nil {
		return nil, fmt.Errorf("Error creating the source connector.\n%v", err)
	}

	// This is a bit lame but without doing the cool stuff with CommandId etc
	// it's the easiest way to make sure the topic exists before continuing
	time.Sleep(5 * time.Second)

	// Create the DOGS stream
	if err := client.Execute(`
	CREATE STREAM DOGS (ID STRING KEY, 
						NAME STRING, 
						DOGSIZE STRING, 
						AGE STRING) 
				  WITH (KAFKA_TOPIC='dogs', 
				  VALUE_FORMAT='JSON');
	`); err != nil {
		return nil, fmt.Errorf("Error creating the DOGS stream.\n%v", err)
	}

	// This is a bit lame but without doing the cool stuff with CommandId etc
	// it's the easiest way to make sure the stream exists before continuing
	time.Sleep(5 * time.Second)

	// Create the DOGS_BY_SIZE table
	if err := client.Execute(`
	CREATE TABLE DOGS_BY_SIZE AS 
		SELECT DOGSIZE AS DOG_SIZE, COUNT(*) AS DOGS_CT 
		FROM DOGS WINDOW TUMBLING (SIZE 15 MINUTE) 
		GROUP BY DOGSIZE;
	`); err != nil {
		return nil, fmt.Errorf("Error creating the DOGS stream.\n%v", err)
	}
	// This is a bit lame but without doing the cool stuff with CommandId etc
	// it's the easiest way to make sure the table exists before continuing
	time.Sleep(10 * time.Second)

	return client, nil
}
