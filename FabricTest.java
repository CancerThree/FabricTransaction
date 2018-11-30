package org.hyperledger.fabric.sdkintegration;

import static java.lang.String.format;
import static org.hyperledger.fabric.sdk.Channel.PeerOptions.createPeerOptions;
import static org.junit.Assert.assertFalse;
import static org.junit.Assert.assertNotNull;

import java.io.File;
import java.nio.file.Paths;
import java.util.Collection;
import java.util.EnumSet;
import java.util.HashMap;
import java.util.LinkedList;
import java.util.Properties;
import java.util.concurrent.TimeUnit;

import org.hyperledger.fabric.sdk.Channel;
import org.hyperledger.fabric.sdk.ChannelConfiguration;
import org.hyperledger.fabric.sdk.EventHub;
import org.hyperledger.fabric.sdk.HFClient;
import org.hyperledger.fabric.sdk.Orderer;
import org.hyperledger.fabric.sdk.Peer;
import org.hyperledger.fabric.sdk.ProposalResponse;
import org.hyperledger.fabric.sdk.TransactionProposalRequest;
import org.hyperledger.fabric.sdk.exception.InvalidArgumentException;
import org.hyperledger.fabric.sdk.exception.ProposalException;
import org.hyperledger.fabric.sdk.Peer.PeerRole;
import org.hyperledger.fabric.sdk.security.CryptoSuite;
import org.hyperledger.fabric.sdk.testutils.TestConfig;
import org.hyperledger.fabric_ca.sdk.HFCAClient;
import org.hyperledger.fabric_ca.sdk.HFCAInfo;

public class FabricTest {

//	public static void invoke() {
//        TransactionProposalRequest transactionProposalRequest = client.newTransactionProposalRequest();
//        transactionProposalRequest.setChaincodeID(chaincodeID);
//        transactionProposalRequest.setChaincodeLanguage(CHAIN_CODE_LANG);
//        //transactionProposalRequest.setFcn("invoke");
//        transactionProposalRequest.setFcn("move");
//        transactionProposalRequest.setProposalWaitTime(testConfig.getProposalWaitTime());
//        transactionProposalRequest.setArgs("a", "b", "100");
//
//        Map<String, byte[]> tm2 = new HashMap<>();
////        tm2.put("HyperLedgerFabric", "TransactionProposalRequest:JavaSDK".getBytes(UTF_8)); //Just some extra junk in transient map
////        tm2.put("method", "TransactionProposalRequest".getBytes(UTF_8)); // ditto
////        tm2.put("result", ":)".getBytes(UTF_8));  // This should be returned see chaincode why.
////        tm2.put(EXPECTED_EVENT_NAME, EXPECTED_EVENT_DATA);  //This should trigger an event see chaincode why.
//
//        transactionProposalRequest.setTransientMap(tm2);
//
//        System.out.println("sending transactionProposal to all peers with arguments: move(a,b,100)");
//
//        //  Collection<ProposalResponse> transactionPropResp = channel.sendTransactionProposalToEndorsers(transactionProposalRequest);
//        Collection<ProposalResponse> transactionPropResp = channel.sendTransactionProposal(transactionProposalRequest, channel.getPeers());
//        for (ProposalResponse response : transactionPropResp) {
//            if (response.getStatus() == ProposalResponse.Status.SUCCESS) {
//            	System.out.printf("Successful transaction proposal response Txid: %s from peer %s\n", response.getTransactionID(), response.getPeer().getName());
//                successful.add(response);
//            } else {
//                failed.add(response);
//            }
//        }
//	}
	
//	private static final TestConfig testConfig = TestConfig.getConfig();
	public static void main(String[] args) throws Exception {
        HFClient client = HFClient.createNewInstance();
        client.setCryptoSuite(CryptoSuite.Factory.getCryptoSuite());
//        SampleUser user = new SampleUser();

        File sampleStoreFile = new File(System.getProperty("java.io.tmpdir") + "/HFCSampletest.properties");
        SampleStore sampleStore = new SampleStore(sampleStoreFile);
//        SampleOrg sampleOrg = testConfig.getIntegrationTestsSampleOrg("peerOrg1");
        SampleOrg sampleOrg = new SampleOrg("tebon", "Tebon");
        sampleOrg.addOrdererLocation("orderer.tebon.com", "grpcs://99.13.43.6:7050");
        sampleOrg.addPeerLocation("peer0.tebon.com", "grpcs://99.13.43.6:7051");
        sampleOrg.setDomainName("tebon.com");
        sampleOrg.addEventHubLocation("peer0.tebon.com", "grpcs://99.13.43.6:7053");
        sampleOrg.setCALocation("http://ca.tebon.com:7054");
        sampleOrg.setCAName("tebon");	//?????
        Properties properties = new Properties();
        String cert = "crypto-config/peerOrganizations/tebon.com/ca/ca.tebon.com-cert.pem";
        File cf = new File(cert);
        if (!cf.exists() || !cf.isFile()) {
            throw new RuntimeException("TEST is missing cert file " + cf.getAbsolutePath());
        }
        properties.setProperty("pemFile", cf.getAbsolutePath());
        properties.setProperty("allowAllHostNames", "false"); //testing environment only NOT FOR PRODUCTION!
        sampleOrg.setCAProperties(properties);
        
        
        
        sampleOrg.setCAClient(HFCAClient.createNewInstance(sampleOrg.getCALocation(), sampleOrg.getCAProperties()));
        HFCAClient ca = sampleOrg.getCAClient();
        ca.setCryptoSuite(CryptoSuite.Factory.getCryptoSuite());
//        sampleOrg.setAdmin("Admin");
        
        HFCAInfo info = ca.info(); //just check if we connect at all.
        
        SampleUser admin = sampleStore.getMember("admin", "tebon");
        if (!admin.isEnrolled()) {  //Preregistered admin only needs to be enrolled with Fabric caClient.
            admin.setEnrollment(ca.enroll("admin", "adminpw"));
            admin.setMspId("Tebon");
        }
        sampleOrg.setAdmin(admin);
        
        final String sampleOrgName = sampleOrg.getName();
        final String sampleOrgDomainName = sampleOrg.getDomainName();
        
        SampleUser peerOrgAdmin = sampleStore.getMember("admin", "tebon", "Tebon",
                Paths.get("crypto-config/peerOrganizations/", sampleOrgDomainName,
                		format("/users/Admin@%s/msp/keystore/Admin@%s.key", sampleOrgDomainName, sampleOrgDomainName)).toFile(),
                Paths.get("crypto-config/peerOrganizations/", sampleOrgDomainName,
                        format("/users/Admin@%s/msp/signcerts/Admin@%s-cert.pem", sampleOrgDomainName, sampleOrgDomainName)).toFile());
        sampleOrg.setPeerAdmin(peerOrgAdmin);
        
        Channel channel = constructChannel("cwjtestcc", client, sampleOrg);
////        sampleStore.saveChannel(fooChannel);
//		// TODO Auto-generated method stub
//		final String channelName = channel.getName();
//        BlockchainInfo channelInfo = channel.queryBlockchainInfo();
//        System.out.println("Channel info for : " + channelName);
//        System.out.println("Channel height: " + channelInfo.getHeight());
//        String chainCurrentHash = Hex.encodeHexString(channelInfo.getCurrentBlockHash());
//        String chainPreviousHash = Hex.encodeHexString(channelInfo.getPreviousBlockHash());
//        System.out.println("Chain current block hash: " + chainCurrentHash);
//        System.out.println("Chainl previous block hash: " + chainPreviousHash);
	}
	
	
    private static String getDomainName(final String name) {
        int dot = name.indexOf(".");
        if (-1 == dot) {
            return null;
        } else {
            return name.substring(dot + 1);
        }

    }
    
	public static Properties getEndPointProperties(final String type, final String name) {
        Properties ret = new Properties();

        final String domainName = getDomainName(name);

        File cert = Paths.get("crypto-config/ordererOrganizations".replace("orderer", type), domainName, type + "s",
                name, "tls/server.crt").toFile();
        if (!cert.exists()) {
            throw new RuntimeException(String.format("Missing cert file for: %s. Could not find at location: %s", name,
                    cert.getAbsolutePath()));
        }

        if ( true/*!isRunningAgainstFabric10()*/) {
            File clientCert;
            File clientKey;
            if ("orderer".equals(type)) {
                clientCert = Paths.get("crypto-config/ordererOrganizations/tebon.com/users/Admin@tebon.com/tls/server.crt").toFile();

                clientKey = Paths.get("crypto-config/ordererOrganizations/tebon.com/users/Admin@tebon.com/tls/server.key").toFile();
            } else {
                clientCert = Paths.get("crypto-config/peerOrganizations/", domainName, "users/Admin@" + domainName, "tls/server.crt").toFile();
                clientKey = Paths.get("crypto-config/peerOrganizations/", domainName, "users/Admin@" + domainName, "tls/server.key").toFile();
            }

            if (!clientCert.exists()) {
                throw new RuntimeException(String.format("Missing  client cert file for: %s. Could not find at location: %s", name,
                        clientCert.getAbsolutePath()));
            }

            if (!clientKey.exists()) {
                throw new RuntimeException(String.format("Missing  client key file for: %s. Could not find at location: %s", name,
                        clientKey.getAbsolutePath()));
            }
            ret.setProperty("clientCertFile", clientCert.getAbsolutePath());
            ret.setProperty("clientKeyFile", clientKey.getAbsolutePath());
        }

        ret.setProperty("pemFile", cert.getAbsolutePath());

        ret.setProperty("hostnameOverride", name);
        ret.setProperty("sslProvider", "openSSL");
        ret.setProperty("negotiationType", "TLS");

        return ret;
    }
	
    static Channel constructChannel(String name, HFClient client, SampleOrg sampleOrg) throws Exception {
        ////////////////////////////
        //Construct the channel
        //

        boolean doPeerEventing = false;
//        boolean doPeerEventing = !testConfig.isRunningAgainstFabric10() && BAR_CHANNEL_NAME.equals(name);
//        boolean doPeerEventing = !testConfig.isRunningAgainstFabric10() && FOO_CHANNEL_NAME.equals(name);
        //Only peer Admin org
        client.setUserContext(sampleOrg.getPeerAdmin());

        Collection<Orderer> orderers = new LinkedList<>();

        for (String orderName : sampleOrg.getOrdererNames()) {

            Properties ordererProperties = getEndPointProperties("orderer",orderName);

            //example of setting keepAlive to avoid timeouts on inactive http2 connections.
            // Under 5 minutes would require changes to server side to accept faster ping rates.
            ordererProperties.put("grpc.NettyChannelBuilderOption.keepAliveTime", new Object[] {5L, TimeUnit.MINUTES});
            ordererProperties.put("grpc.NettyChannelBuilderOption.keepAliveTimeout", new Object[] {8L, TimeUnit.SECONDS});
            ordererProperties.put("grpc.NettyChannelBuilderOption.keepAliveWithoutCalls", new Object[] {true});

            orderers.add(client.newOrderer(orderName, sampleOrg.getOrdererLocation(orderName),
                    ordererProperties));
        }
        Channel newChannel = client.newChannel(name);
        
        for (Orderer orderer : orderers) { //add remaining orderers if any.
            newChannel.addOrderer(orderer);
        }

        //Just pick the first orderer in the list to create the channel.

//        Orderer anOrderer = orderers.iterator().next();
//        orderers.remove(anOrderer);

//        ChannelConfiguration channelConfiguration = new ChannelConfiguration(new File("/sdkintegration/e2e-2Orgs/" + TestConfig.FAB_CONFIG_GEN_VERS + "/" + name + ".tx"));

        //Create channel that has only one signer that is this orgs peer admin. If channel creation policy needed more signature they would need to be added too.
//        Channel newChannel = client.newChannel(name, anOrderer, channelConfiguration, client.getChannelConfigurationSignature(channelConfiguration, sampleOrg.getPeerAdmin()));
//        Channel newChannel = Channel.createNewInstance(name, client);
//        Channel newChannel = client.newChannel(name, anOrderer, channelConfiguration, client.getChannelConfigurationSignature(channelConfiguration, sampleOrg.getPeerAdmin()));
//        System.out.printf("Created channel %s\n", name);
//        newChannel.addOrderer();
        boolean everyother = true; //test with both cases when doing peer eventing.
        for (String peerName : sampleOrg.getPeerNames()) {
            String peerLocation = sampleOrg.getPeerLocation(peerName);

            Properties peerProperties = getEndPointProperties("peer",peerName); //test properties for peer.. if any.
            if (peerProperties == null) {
                peerProperties = new Properties();
            }

            //Example of setting specific options on grpc's NettyChannelBuilder
            peerProperties.put("grpc.NettyChannelBuilderOption.maxInboundMessageSize", 9000000);

            Peer peer = client.newPeer(peerName, peerLocation, peerProperties);
            if (doPeerEventing && everyother) {
                newChannel.joinPeer(peer, createPeerOptions().setPeerRoles(EnumSet.of(PeerRole.ENDORSING_PEER, PeerRole.LEDGER_QUERY, PeerRole.CHAINCODE_QUERY, PeerRole.EVENT_SOURCE))); //Default is all roles.
            } else {
                // Set peer to not be all roles but eventing.
                newChannel.joinPeer(peer, createPeerOptions().setPeerRoles(EnumSet.of(PeerRole.ENDORSING_PEER, PeerRole.LEDGER_QUERY, PeerRole.CHAINCODE_QUERY)));
            }
            System.out.printf("Peer %s joined channel %s\n", peerName, name);
            everyother = !everyother;
        }
        //just for testing ...
        if (doPeerEventing) {
            // Make sure there is one of each type peer at the very least.
            assertFalse(newChannel.getPeers(EnumSet.of(PeerRole.EVENT_SOURCE)).isEmpty());
            assertFalse(newChannel.getPeers(PeerRole.NO_EVENT_SOURCE).isEmpty());
        }

        

        for (String eventHubName : sampleOrg.getEventHubNames()) {

            final Properties eventHubProperties = getEndPointProperties("peer", eventHubName);

            eventHubProperties.put("grpc.NettyChannelBuilderOption.keepAliveTime", new Object[] {5L, TimeUnit.MINUTES});
            eventHubProperties.put("grpc.NettyChannelBuilderOption.keepAliveTimeout", new Object[] {8L, TimeUnit.SECONDS});

            EventHub eventHub = client.newEventHub(eventHubName, sampleOrg.getEventHubLocation(eventHubName),
                    eventHubProperties);
            newChannel.addEventHub(eventHub);
        }

        return newChannel.initialize();

//        System.out.printf("Finished initialization channel %s\n", name);
//
//        //Just checks if channel can be serialized and deserialized .. otherwise this is just a waste :)
//        byte[] serializedChannelBytes = newChannel.serializeChannel();
//        newChannel.shutdown(true);
//
//        return client.deSerializeChannel(serializedChannelBytes).initialize();

    }

}
