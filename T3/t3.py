from redis import Redis
from rq import Queue
import sys
from functions_hash import *

redis_conn = Redis(host="localhost", port="6379")

q_low = Queue("low", connection=redis_conn)
q_med = Queue("medium", connection=redis_conn)
q_high = Queue("high", connection=redis_conn)


filename = str(sys.argv[1])
file1 = open(filename, 'r') 
Lines = file1.readlines() 

file1.close() 

f = open("output.txt", "w")
index = 0


for line in Lines: 
    line = line.strip("\n")
    client, hash_string, k, end_of_hash = line.split(" ")
    k = int(k)

    if client[0] == "C":
        q_low.enqueue(proof_of_work_with_client, index, client, k, hash_string, end_of_hash)
    elif client[0] == "T":
        q_med.enqueue(proof_of_work_with_client, index, client, k, hash_string, end_of_hash)
    else:
        q_high.enqueue(proof_of_work_with_client, index, client, k, hash_string, end_of_hash)

    index += 1




