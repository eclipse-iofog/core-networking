package main.com.iotracks.core_networking.local_client;

import io.netty.channel.Channel;
import main.com.iotracks.core_networking.main.CoreNetworking;
import main.com.iotracks.core_networking.local_client.private_client.PrivateLocalClient;
import main.com.iotracks.core_networking.local_client.public_client.PublicLocalClient;

/**
 * Created by saeid on 4/13/16.
 */
public class LocalClientBuilder {
    public static LocalClient build(Channel comSatChannel) {
        if (CoreNetworking.config.getMode().equals("public"))
            return new PublicLocalClient(comSatChannel);
        else
            return new PrivateLocalClient(comSatChannel);
    }
}
