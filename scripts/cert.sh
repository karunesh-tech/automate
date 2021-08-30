#!/bin/bash
 
#First, create a private key for the CA: 
openssl genrsa -out MyRootCA.key 2048
 
#Create the CA and enter the Organization details: 
openssl req -x509 -new -key MyRootCA.key -sha256 -out MyRootCA.pem -subj '/C=US/ST=Washington/L=Seattle/O=Chef Software Inc/CN=chefrootca'
 
#the rsa keys 
openssl genrsa -out admin-pkcs12.key 2048
 
#IMPORTANT: Convert these to PKCS#5 v1.5 to work correctly with the JDK.
openssl pkcs8 -v1 "PBE-SHA1-3DES" -in "admin-pkcs12.key" -topk8 -out "admin.key" -nocrypt

#Create the CSR and enter the organization and server details for the node key 
openssl req -new -key admin.key -out admin.csr -subj '/C=US/ST=Washington/L=Seattle/O=Chef Software Inc/CN=chefadmin'
 
#Use the CSR to generate the signed node Certificate:
openssl x509 -req -in admin.csr -CA MyRootCA.pem -CAkey MyRootCA.key -CAcreateserial -out admin.pem -sha256
 

# root pem cert that signed the below cert/key pairs below
# Used for hab_sup_http_gateway_ca_cert 
cat <<EOF >> ../terraform/a2ha-terraform/variables_common.tf

variable "hab_sup_http_gateway_ca_cert" {
  default = <<CERT
$(cat MyRootCA.key)
CERT


  description = "Issuer of the TLS cert used for the HTTP gateway in PEM format."
}

variable "hab_sup_http_gateway_ca_cert" {
  default = <<CERT
$(cat admin.pem)
CERT


  description = "Issuer of the TLS cert used for the HTTP gateway in PEM format."
}

variable "hab_sup_http_gateway_ca_cert" {
  default = <<CERT
$(cat admin.key)
CERT


  description = "Issuer of the TLS cert used for the HTTP gateway in PEM format."
}

EOF
 
rm MyRootCA.key MyRootCA.pem MyRootCA.srl admin-pkcs12.key admin.csr admin.key admin.pem