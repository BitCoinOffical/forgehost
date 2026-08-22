resource "kafka_topic" "name" {
    name = ""
    partitions = 4
    replication_factor = 1

    config = {
        "retention.ms" = 7 * 24 * 60 * 60 * 10000 //7 days
    }
}