package main.org.eclipse.iofog.core_networking.comsat_client;

import io.netty.channel.ChannelHandlerContext;
import io.netty.channel.SimpleChannelInboundHandler;
import main.org.eclipse.iofog.core_networking.local_client.LocalClient;
import main.org.eclipse.iofog.core_networking.local_client.LocalClientBuilder;

import java.util.logging.Logger;

/**
 * ComSat connection client handler
 * <p>
 * Created by saeid on 4/8/16.
 */
public class ComSatClientHandler extends SimpleChannelInboundHandler<byte[]> {

    private final Logger log = Logger.getLogger(ComSatClientHandler.class.getName());
    private final ComSatClient client;
    private LocalClient localClient = null;

    public ComSatClientHandler(ComSatClient client) {
        this.client = client;
    }

    /**
     * creates private/public local client once connection to ComSat server established
     *
     * @param ctx
     * @throws Exception
     */
    @Override
    public void channelActive(ChannelHandlerContext ctx) throws Exception {
        if (localClient == null)
            localClient = LocalClientBuilder.build(client.getChannel());
    }

    /**
     * disconnects from ComSat server when connection to local container been lost.
     *
     * @param ctx
     * @throws Exception
     */
    @Override
    public void channelInactive(ChannelHandlerContext ctx) throws Exception {
        if (localClient != null)
            localClient.closeConnection();

        ctx.close();
    }

    /**
     * receives data from ComSat server.
     * If data is equal to "BEAT", "AUTHORIZED" or "BEATBEAT", updates last seen value of the connection
     * otherwise pipes received data to local container
     *
     * @param ctx
     * @param bytes
     * @throws Exception
     */
    @Override
    protected void channelRead0(ChannelHandlerContext ctx, byte[] bytes) throws Exception {
        if (bytes.length < 11) {
            String contentString = new String(bytes);
            log.info(String.format("#%d : %s", client.getConnectionId(), contentString));
            if (contentString.equals("BEAT") || contentString.equals("AUTHORIZED") || contentString.equals("BEATBEAT")) {
                client.seen();
                return;
            }
        }

        log.info(String.format("#%d : %d bytes received from comsat", client.getConnectionId(), bytes.length));
        if (!localClient.isConnected()) {
            log.info(String.format("#%d : connecting to client", client.getConnectionId()));
            if (!localClient.connect(2000)) {
                log.warning(String.format("#%d : unable to connect to local client", client.getConnectionId()));
                return;
            }
        }

        log.info(String.format("#%d : piping bytes from comsat to client", client.getConnectionId()));
        localClient.handleMessage(bytes);
    }

    @Override
    public void exceptionCaught(ChannelHandlerContext ctx, Throwable cause) {
        log.warning(String.format("#%d : %s", client.getConnectionId(), cause.getMessage()));
        ctx.close();
    }

}
