# allora-worker

## Install requirements
```
sudo apt update && sudo apt upgrade -y
sudo apt install jq -y

# install docker
curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo gpg --dearmor -o /usr/share/keyrings/docker-archive-keyring.gpg

echo "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/docker-archive-keyring.gpg] https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable" | sudo tee /etc/apt/sources.list.d/docker.list > /dev/null

sudo apt-get update
sudo apt-get install docker-ce docker-ce-cli containerd.io
docker version

# install docker-compose
VER=$(curl -s https://api.github.com/repos/docker/compose/releases/latest | grep tag_name | cut -d '"' -f 4)

curl -L "https://github.com/docker/compose/releases/download/"$VER"/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose

chmod +x /usr/local/bin/docker-compose
docker-compose --version
```


request some faucet from the [Allora Testnet Faucet](https://faucet.testnet-1.testnet.allora.network/) 

## Stop the old version
If you've previously run the old version and want to stop it before proceeding, follow these commands
```
docker stop custom-inference
docker stop custom-worker
docker container prune -f
```

## Run the custom model
1. Create an account and obtain an Upshot ApiKey [here](https://developer.upshot.xyz)
2. create an account and optain [here](https://www.coingecko.com/en/developers/dashboard)
3. 
4. Clone the git repository
```
git clone https://github.com/sarox0987/allora-worker.git
cd allora-worker
```

3. Run the bash script
```
bash run.sh
```
