import hashlib

def validate_proof_of_work(last_k, last_hash, k, end_hash):
	sha = hashlib.sha256(f'{last_k}{last_hash}{k}'.encode())
	return sha.hexdigest()[:len(end_hash)] == end_hash


def generate_proof_of_work(last_k, last_hash, end_hash):
	k = 0
	while not validate_proof_of_work(last_k, last_hash, k, end_hash):
		k += 1

	return k

def proof_of_work_with_client(index, client, last_k, last_hash, end_hash):
	f = open("output.txt", "a")
	result = generate_proof_of_work(last_k, last_hash, end_hash)
	string_to_write = f"{index} {client} {result}\n"
	f.write(string_to_write)
	f.close()

